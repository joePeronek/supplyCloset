package main

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLoadData(t *testing.T) {
	resetState()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	db = mockDB

	invRows := sqlmock.NewRows([]string{"id", "uniform_type", "gender", "name", "style", "size", "quantity"}).
		AddRow(1, "", "", "Pen", "", "", 3)
	mock.ExpectQuery("SELECT id, uniform_type, gender, name, style, size, quantity FROM inventory").
		WillReturnRows(invRows)

	issRows := sqlmock.NewRows([]string{"item_id", "item_name", "person", "issued_by", "issued_at"}).
		AddRow(1, "Pen", "Alice", "Bob", time.Now())
	mock.ExpectQuery("SELECT item_id, item_name, person, issued_by, issued_at FROM issued").
		WillReturnRows(issRows)

	if err := loadData(); err != nil {
		t.Fatalf("loadData: %v", err)
	}
	if len(items) != 1 || items[1].Name != "Pen" {
		t.Fatalf("items not loaded: %+v", items)
	}
	if len(issued) != 1 || issued[0].Person != "Alice" {
		t.Fatalf("issued not loaded: %+v", issued)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPopulateDB(t *testing.T) {
	resetState()
	items[1] = &InventoryItem{ID: 1, Name: "Pen", Quantity: 4}
	issued = append(issued, IssuedItem{ItemID: 1, ItemName: "Pen", Person: "Alice", IssuedBy: "Bob", IssuedAt: time.Now()})

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	db = mockDB

	mock.ExpectExec("INSERT INTO inventory").
		WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg(), "Pen", sqlmock.AnyArg(), sqlmock.AnyArg(), 4).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO issued").
		WithArgs(1, "Pen", "Alice", "Bob", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := populateDB(); err != nil {
		t.Fatalf("populateDB: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
