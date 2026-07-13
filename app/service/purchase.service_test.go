package service

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/model"
	"github.com/jerobas/saas/repository"
)

func TestBuyPostsPurchaseAndUpdatesStock(t *testing.T) {
	db := testDatabase(t)
	itemID, err := repository.NewItemRepository(db.Conn).Create(&model.ItemInsertDTO{
		Name:             "Farinha",
		Unit:             "kg",
		Purchasable:      1,
		DefaultSalePrice: sql.NullInt64{},
	})
	if err != nil {
		t.Fatal(err)
	}

	eventID, err := NewPurchaseService(db).Buy(CreatePurchaseInput{
		Lines: []PurchaseLineInput{{ItemID: itemID, Quantity: 2.5, UnitCost: 1200}},
	})
	if err != nil {
		t.Fatal(err)
	}
	purchase, err := NewPurchaseService(db).GetPurchaseByID(eventID)
	if err != nil {
		t.Fatal(err)
	}
	if purchase.Status != "POSTED" || len(purchase.PurchaseLines) != 1 {
		t.Fatalf("unexpected purchase: status=%s lines=%d", purchase.Status, len(purchase.PurchaseLines))
	}

	stock, err := repository.NewItemStockRepository(db.Conn).GetByID(itemID)
	if err != nil {
		t.Fatal(err)
	}
	if stock.Quantity != 2.5 || !stock.AverageUnitCost.Valid || stock.AverageUnitCost.Int64 != 1200 {
		t.Fatalf("unexpected stock: %+v", stock)
	}
}

func TestBuyRollsBackOnInvalidItem(t *testing.T) {
	db := testDatabase(t)
	_, err := NewPurchaseService(db).Buy(CreatePurchaseInput{
		Lines: []PurchaseLineInput{{ItemID: 999, Quantity: 1, UnitCost: 100}},
	})
	if err == nil {
		t.Fatal("expected invalid item error")
	}
	var events int
	if err := db.Conn.QueryRow("SELECT COUNT(*) FROM events").Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 0 {
		t.Fatalf("event count = %d, want rollback to leave 0", events)
	}
}

func testDatabase(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewDatabase(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
