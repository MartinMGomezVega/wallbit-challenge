package invoice

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func TestInvoiceKeyAndTypes(t *testing.T) {
	rate := money.Rate(14861100)
	inv := Invoice{
		ID:           "inv_1",
		ClientID:     "cli_123",
		Type:         TypeE,
		PointOfSale:  2,
		Number:       45,
		IssueDate:    time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		Currency:     money.USD,
		NetAmount:    money.New(200000, money.USD),
		VatAmount:    money.Zero(money.USD),
		TotalAmount:  money.New(200000, money.USD),
		FXRate:       &rate,
		ReceiverName: "Northwind Labs LLC",
		ReceiverDoc:  "12-34567890",
		Source:       SourceCSV,
		CreatedAt:    time.Now(),
	}

	key := inv.Key()
	expectedKey := Key{
		ClientID:    "cli_123",
		Type:        TypeE,
		PointOfSale: 2,
		Number:      45,
	}
	if key != expectedKey {
		t.Fatalf("Key = %+v, want %+v", key, expectedKey)
	}

	expectedStr := "cli_123-E-0002-00000045"
	if key.String() != expectedStr {
		t.Fatalf("Key.String() = %s, want %s", key.String(), expectedStr)
	}

	// PurchaseInvoice
	pi := PurchaseInvoice{
		ID:           "pi_1",
		ClientID:     "cli_123",
		IssueDate:    time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC),
		SupplierName: "Distribuidora Tech SRL",
		SupplierCUIT: "30-71234567-8",
		NetAmount:    money.New(10000000, money.ARS),
		VatAmount:    money.New(2100000, money.ARS),
	}
	if pi.SupplierName != "Distribuidora Tech SRL" || pi.NetAmount.Amount != 10000000 {
		t.Fatalf("unexpected PurchaseInvoice: %+v", pi)
	}
}
