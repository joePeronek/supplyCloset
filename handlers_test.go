package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestInventoryHandlerPostDBAutoID(t *testing.T) {
	resetState()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	db = mockDB

	mock.ExpectQuery("INSERT INTO inventory").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Marker", sqlmock.AnyArg(), sqlmock.AnyArg(), 5).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	body := strings.NewReader(`{"name":"Marker","quantity":5}`)
	req := httptest.NewRequest(http.MethodPost, "/inventory", body)
	rr := httptest.NewRecorder()

	inventoryHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if items[42] == nil || items[42].Name != "Marker" {
		t.Fatalf("item not added with generated id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestInventoryHandlerPostDBWithID(t *testing.T) {
	resetState()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	db = mockDB

	mock.ExpectExec("INSERT INTO inventory").
		WithArgs(5, sqlmock.AnyArg(), sqlmock.AnyArg(), "Marker", sqlmock.AnyArg(), sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := strings.NewReader(`{"id":5,"name":"Marker","quantity":2}`)
	req := httptest.NewRequest(http.MethodPost, "/inventory", body)
	rr := httptest.NewRecorder()

	inventoryHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if items[5] == nil || items[5].Quantity != 2 {
		t.Fatalf("item not stored with provided id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
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
