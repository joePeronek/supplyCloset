package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB(conn string) error {
	var err error
	db, err = sql.Open("postgres", conn)
	if err != nil {
		return err
	}
	// create new inventory table if it doesn't exist
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS inventory (
        codex/modify-inventory-table-and-update-handlers
        id INT PRIMARY KEY,
        uniform_type TEXT,
        gender TEXT,
        item TEXT,
        style TEXT,
        size TEXT,
         main
        quantity INT NOT NULL
    )`); err != nil {
		return err
	}
	// handle upgrades from the old schema
	// rename legacy "name" column to "item" if present
	db.Exec(`ALTER TABLE inventory RENAME COLUMN name TO item`)
	// add any missing columns
	alterCols := []string{
		"ALTER TABLE inventory ADD COLUMN IF NOT EXISTS uniform_type TEXT",
		"ALTER TABLE inventory ADD COLUMN IF NOT EXISTS gender TEXT",
		"ALTER TABLE inventory ADD COLUMN IF NOT EXISTS style TEXT",
		"ALTER TABLE inventory ADD COLUMN IF NOT EXISTS size TEXT",
		"ALTER TABLE inventory ADD COLUMN IF NOT EXISTS item TEXT",
	}
	for _, stmt := range alterCols {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS issued (
        id SERIAL PRIMARY KEY,
        item_id INT NOT NULL,
        item_name TEXT NOT NULL,
        person TEXT NOT NULL,
        issued_by TEXT NOT NULL,
        issued_at TIMESTAMPTZ NOT NULL
    )`); err != nil {
		return err
	}
	return nil
}

func loadData() error {
	invRows, err := db.Query(`SELECT id, uniform_type, gender, item, style, size, quantity FROM inventory`)
	if err != nil {
		return err
	}
	defer invRows.Close()
	mu.Lock()
	for invRows.Next() {
		var it InventoryItem
		if err := invRows.Scan(&it.ID, &it.UniformType, &it.Gender, &it.Item, &it.Style, &it.Size, &it.Quantity); err != nil {
			mu.Unlock()
			return err
		}
		items[it.ID] = &it
	}
	mu.Unlock()

	issRows, err := db.Query(`SELECT item_id, item_name, person, issued_by, issued_at FROM issued`)
	if err != nil {
		return err
	}
	defer issRows.Close()
	mu.Lock()
	for issRows.Next() {
		var iss IssuedItem
		if err := issRows.Scan(&iss.ItemID, &iss.ItemName, &iss.Person, &iss.IssuedBy, &iss.IssuedAt); err != nil {
			mu.Unlock()
			return err
		}
		issued = append(issued, iss)
	}
	mu.Unlock()
	return nil
}

func populateDB() error {
	for _, it := range items {
		if _, err := db.Exec(`INSERT INTO inventory (id, uniform_type, gender, item, style, size, quantity) VALUES ($1, $2, $3, $4, $5, $6, $7)
            ON CONFLICT (id) DO UPDATE SET uniform_type=EXCLUDED.uniform_type, gender=EXCLUDED.gender, item=EXCLUDED.item, style=EXCLUDED.style, size=EXCLUDED.size, quantity=EXCLUDED.quantity`,
			it.ID, it.UniformType, it.Gender, it.Item, it.Style, it.Size, it.Quantity); err != nil {
			return err
		}
	}
	for _, iss := range issued {
		if _, err := db.Exec(`INSERT INTO issued (item_id, item_name, person, issued_by, issued_at)
            VALUES ($1, $2, $3, $4, $5)`,
			iss.ItemID, iss.ItemName, iss.Person, iss.IssuedBy, iss.IssuedAt); err != nil {
			return err
		}
	}
	return nil
}
