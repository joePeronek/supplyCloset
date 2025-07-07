package main

import (
	"encoding/json"
	"net/http"
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
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("client")))
	http.HandleFunc("/inventory", inventoryHandler)
	http.HandleFunc("/issue", issueHandler)
	http.HandleFunc("/issued", issuedHandler)
	http.ListenAndServe(":8080", nil)
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
		items[it.ID] = &it
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
