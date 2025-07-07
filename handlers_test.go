package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func resetState() {
	items = map[int]*InventoryItem{}
	issued = []IssuedItem{}
	nextID = 1
	db = nil
}

func TestInventoryHandlerGet(t *testing.T) {
	resetState()
	items[1] = &InventoryItem{ID: 1, Name: "Pen", Quantity: 10}

	req := httptest.NewRequest(http.MethodGet, "/inventory", nil)
	rr := httptest.NewRecorder()

	inventoryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp []InventoryItem
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].Name != "Pen" {
		t.Fatalf("unexpected body: %+v", resp)
	}
}

func TestInventoryHandlerPost(t *testing.T) {
	resetState()

	body := strings.NewReader(`{"name":"Marker","quantity":5}`)
	req := httptest.NewRequest(http.MethodPost, "/inventory", body)
	rr := httptest.NewRecorder()

	inventoryHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if len(items) != 1 {
		t.Fatalf("item not added")
	}
	it := items[1]
	if it.Name != "Marker" || it.Quantity != 5 {
		t.Fatalf("unexpected item: %+v", it)
	}
}

func TestIssueHandler(t *testing.T) {
	resetState()
	items[1] = &InventoryItem{ID: 1, Name: "Pen", Quantity: 2}

	body := strings.NewReader(`{"itemId":1,"person":"Alice","issuedBy":"Bob"}`)
	req := httptest.NewRequest(http.MethodPost, "/issue", body)
	rr := httptest.NewRecorder()

	issueHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if items[1].Quantity != 1 {
		t.Fatalf("quantity not decremented: %d", items[1].Quantity)
	}
	if len(issued) != 1 || issued[0].Person != "Alice" {
		t.Fatalf("issued record not added")
	}
}

func TestIssuedHandler(t *testing.T) {
	resetState()
	issued = append(issued, IssuedItem{ItemID: 1, ItemName: "Pen", Person: "Alice", IssuedBy: "Bob", IssuedAt: time.Now()})

	req := httptest.NewRequest(http.MethodGet, "/issued", nil)
	rr := httptest.NewRecorder()

	issuedHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp []IssuedItem
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].Person != "Alice" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
