package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type InventoryItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type IssuedItem struct {
	ItemID   int       `json:"itemId"`
	ItemName string    `json:"itemName"`
	Person   string    `json:"person"`
	IssuedBy string    `json:"issuedBy"`
	IssuedAt time.Time `json:"issuedAt"`
}

var (
	items  = map[int]*InventoryItem{}
	issued = []IssuedItem{}
	mu     sync.Mutex
	nextID = 1
)

func main() {
	conn := os.Getenv("DATABASE_URL")
	if conn != "" {
		if err := initDB(conn); err != nil {
			log.Fatal(err)
		}
		if err := loadData(); err != nil {
			log.Fatal(err)
		}
	}

	mu.Lock()
	for id := range items {
		if id >= nextID {
			nextID = id + 1
		}
	}
	mu.Unlock()

	http.Handle("/", http.FileServer(http.Dir("client")))
	http.HandleFunc("/inventory", inventoryHandler)
	http.HandleFunc("/issue", issueHandler)
	http.HandleFunc("/issued", issuedHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func inventoryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		defer mu.Unlock()
		list := make([]InventoryItem, 0, len(items))
		for _, v := range items {
			list = append(list, *v)
		}
		json.NewEncoder(w).Encode(list)
	case http.MethodPost:
		var it InventoryItem
		if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		it.ID = nextID
		nextID++
		items[it.ID] = &it
		if db != nil {
			if _, err := db.Exec(`INSERT INTO inventory (id, name, quantity) VALUES ($1, $2, $3)
                               ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, quantity=EXCLUDED.quantity`,
				it.ID, it.Name, it.Quantity); err != nil {
				mu.Unlock()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func issueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ItemID   int    `json:"itemId"`
		Person   string `json:"person"`
		IssuedBy string `json:"issuedBy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	item, ok := items[req.ItemID]
	if !ok || item.Quantity <= 0 {
		mu.Unlock()
		http.Error(w, "item unavailable", http.StatusBadRequest)
		return
	}
	item.Quantity--
	iss := IssuedItem{
		ItemID:   req.ItemID,
		ItemName: item.Name,
		Person:   req.Person,
		IssuedBy: req.IssuedBy,
		IssuedAt: time.Now(),
	}
	issued = append(issued, iss)
	if db != nil {
		if _, err := db.Exec(`UPDATE inventory SET quantity = quantity - 1 WHERE id = $1`, req.ItemID); err != nil {
			mu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := db.Exec(`INSERT INTO issued (item_id, item_name, person, issued_by, issued_at)
                       VALUES ($1, $2, $3, $4, $5)`,
			req.ItemID, item.Name, req.Person, req.IssuedBy, iss.IssuedAt); err != nil {
			mu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	mu.Unlock()
	w.WriteHeader(http.StatusCreated)
}

func issuedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(issued)
}
