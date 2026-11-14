package reconciliation

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestReconcile(t *testing.T) {
	defaultParams := Params{
		ToleranceBp:      500, // 5% = 500 bp
		WindowBeforeDays: 30,
		WindowAfterDays:  15,
	}

	date10 := time.Date(2026, 10, 10, 12, 0, 0, 0, period.Location())

	t.Run("Match exacto", func(t *testing.T) {
		m := movement.Movement{
			ID:         "mov_1",
			ClientID:   "cli_1",
			Type:       movement.TypeInboundTransfer,
			Amount:     money.New(200000, money.USD), // 2000 USD
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Northwind Labs LLC",
			},
		}
		inv := invoice.Invoice{
			ID:           "inv_1",
			ClientID:     "cli_1",
			Type:         invoice.TypeE,
			Currency:     money.USD,
			TotalAmount:  money.New(200000, money.USD),
			IssueDate:    date10,
			ReceiverName: "Northwind Labs LLC",
		}

		res := Reconcile([]movement.Movement{m}, []invoice.Invoice{inv}, defaultParams)
		if len(res.Matched) != 1 || len(res.UnmatchedIncome) != 0 || len(res.UnmatchedInvoices) != 0 {
			t.Fatalf("unexpected result: %+v", res)
		}
		if res.Matched[0].Movement.ID != "mov_1" || res.Matched[0].Invoice.ID != "inv_1" || res.Matched[0].DiffAmount != 0 {
			t.Fatalf("matched pair mismatch: %+v", res.Matched[0])
		}
	})

	t.Run("Con comisión dentro de tolerancia", func(t *testing.T) {
		// Factura de USD 1000, cobro de USD 970 (3% comisión < 5% tolerancia)
		m := movement.Movement{
			ID:         "mov_1",
			ClientID:   "cli_1",
			Type:       movement.TypeInboundTransfer,
			Amount:     money.New(97000, money.USD),
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Deel",
			},
		}
		inv := invoice.Invoice{
			ID:           "inv_1",
			ClientID:     "cli_1",
			Type:         invoice.TypeE,
			Currency:     money.USD,
			TotalAmount:  money.New(100000, money.USD),
			IssueDate:    date10,
			ReceiverName: "Deel",
		}

		res := Reconcile([]movement.Movement{m}, []invoice.Invoice{inv}, defaultParams)
		if len(res.Matched) != 1 || res.Matched[0].DiffAmount != -3000 {
			t.Fatalf("expected matched with diff -3000, got: %+v", res)
		}
	})

	t.Run("Fuera de tolerancia", func(t *testing.T) {
		// Factura de USD 1000, cobro de USD 900 (10% diferencia > 5% tolerancia)
		m := movement.Movement{
			ID:         "mov_1",
			ClientID:   "cli_1",
			Type:       movement.TypeInboundTransfer,
			Amount:     money.New(90000, money.USD),
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Deel",
			},
		}
		inv := invoice.Invoice{
			ID:           "inv_1",
			ClientID:     "cli_1",
			Type:         invoice.TypeE,
			Currency:     money.USD,
			TotalAmount:  money.New(100000, money.USD),
			IssueDate:    date10,
			ReceiverName: "Deel",
		}

		res := Reconcile([]movement.Movement{m}, []invoice.Invoice{inv}, defaultParams)
		if len(res.Matched) != 0 || len(res.UnmatchedIncome) != 1 || len(res.UnmatchedInvoices) != 1 {
			t.Fatalf("expected unmatched, got: %+v", res)
		}
	})

	t.Run("Cobro anticipado (antes de la factura)", func(t *testing.T) {
		// Cobro el 2026-10-05, factura el 2026-10-15 (10 días de diferencia <= 15 días windowAfter)
		m := movement.Movement{
			ID:         "mov_1",
			ClientID:   "cli_1",
			Type:       movement.TypeInboundTransfer,
			Amount:     money.New(150000, money.USD),
			OccurredAt: time.Date(2026, 10, 5, 10, 0, 0, 0, period.Location()),
			Counterparty: movement.Counterparty{
				Name: "Brightloop GmbH",
			},
		}
		inv := invoice.Invoice{
			ID:           "inv_1",
			ClientID:     "cli_1",
			Type:         invoice.TypeE,
			Currency:     money.USD,
			TotalAmount:  money.New(150000, money.USD),
			IssueDate:    time.Date(2026, 10, 15, 10, 0, 0, 0, period.Location()),
			ReceiverName: "Brightloop GmbH",
		}

		res := Reconcile([]movement.Movement{m}, []invoice.Invoice{inv}, defaultParams)
		if len(res.Matched) != 1 {
			t.Fatalf("expected 1 match for early payment, got: %+v", res)
		}
	})

	t.Run("Nombre con sufijo distinto (Acme Inc. vs ACME)", func(t *testing.T) {
		m := movement.Movement{
			ID:         "mov_1",
			ClientID:   "cli_1",
			Type:       movement.TypeInboundTransfer,
			Amount:     money.New(120000, money.USD),
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Acme, Inc.",
			},
		}
		inv := invoice.Invoice{
			ID:           "inv_1",
			ClientID:     "cli_1",
			Type:         invoice.TypeE,
			Currency:     money.USD,
			TotalAmount:  money.New(120000, money.USD),
			IssueDate:    date10,
			ReceiverName: "ACME",
		}

		res := Reconcile([]movement.Movement{m}, []invoice.Invoice{inv}, defaultParams)
		if len(res.Matched) != 1 {
			t.Fatalf("expected match with normalized suffixes, got: %+v", res)
		}
	})

	t.Run("Dos facturas iguales para dos cobros (matching 1 a 1 sin duplicados)", func(t *testing.T) {
		m1 := movement.Movement{
			ID:         "mov_1",
			Amount:     money.New(100000, money.USD),
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Acme Corp",
			},
		}
		m2 := movement.Movement{
			ID:         "mov_2",
			Amount:     money.New(100000, money.USD),
			OccurredAt: date10.AddDate(0, 0, 2),
			Counterparty: movement.Counterparty{
				Name: "Acme Corp",
			},
		}

		inv1 := invoice.Invoice{
			ID:           "inv_1",
			Currency:     money.USD,
			TotalAmount:  money.New(100000, money.USD),
			IssueDate:    date10,
			ReceiverName: "Acme Corp",
		}
		inv2 := invoice.Invoice{
			ID:           "inv_2",
			Currency:     money.USD,
			TotalAmount:  money.New(100000, money.USD),
			IssueDate:    date10.AddDate(0, 0, 2),
			ReceiverName: "Acme Corp",
		}

		res := Reconcile([]movement.Movement{m1, m2}, []invoice.Invoice{inv1, inv2}, defaultParams)
		if len(res.Matched) != 2 || len(res.UnmatchedIncome) != 0 || len(res.UnmatchedInvoices) != 0 {
			t.Fatalf("expected 2 exact pairs matched 1-to-1, got: %+v", res)
		}
		if res.Matched[0].Invoice.ID == res.Matched[1].Invoice.ID {
			t.Fatal("invoice was matched more than once!")
		}
	})

	t.Run("Sin candidatos", func(t *testing.T) {
		m := movement.Movement{
			ID:         "mov_1",
			Amount:     money.New(50000, money.USD),
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Unknown Client",
			},
		}
		inv := invoice.Invoice{
			ID:           "inv_1",
			Currency:     money.USD,
			TotalAmount:  money.New(50000, money.USD),
			IssueDate:    date10,
			ReceiverName: "Different Company",
		}

		res := Reconcile([]movement.Movement{m}, []invoice.Invoice{inv}, defaultParams)
		if len(res.Matched) != 0 || len(res.UnmatchedIncome) != 1 || len(res.UnmatchedInvoices) != 1 {
			t.Fatalf("expected 0 matches, got: %+v", res)
		}
	})

	t.Run("Empates en orden de movimientos y candidatos", func(t *testing.T) {
		m1 := movement.Movement{
			ID:         "mov_B",
			Amount:     money.New(50000, money.USD),
			OccurredAt: date10,
			Counterparty: movement.Counterparty{
				Name: "Acme Corp",
			},
		}
		m2 := movement.Movement{
			ID:         "mov_A",
			Amount:     money.New(50000, money.USD),
			OccurredAt: date10, // Misma fecha -> orden por ID
			Counterparty: movement.Counterparty{
				Name: "Acme Corp",
			},
		}

		inv1 := invoice.Invoice{
			ID:           "inv_B",
			Currency:     money.USD,
			TotalAmount:  money.New(50000, money.USD),
			IssueDate:    date10,
			ReceiverName: "Acme Corp",
		}
		inv2 := invoice.Invoice{
			ID:           "inv_A",
			Currency:     money.USD,
			TotalAmount:  money.New(50000, money.USD),
			IssueDate:    date10, // Misma fecha y monto -> orden por ID
			ReceiverName: "Acme Corp",
		}

		res := Reconcile([]movement.Movement{m1, m2}, []invoice.Invoice{inv1, inv2}, defaultParams)
		if len(res.Matched) != 2 {
			t.Fatalf("expected 2 matches, got %d", len(res.Matched))
		}
	})
}

func TestKnownPayers(t *testing.T) {
	invoices := []invoice.Invoice{
		{ReceiverName: "Northwind Labs LLC"},
		{ReceiverName: "Brightloop GmbH"},
		{ReceiverName: "Acme Digital, Inc."},
	}

	payers := KnownPayers(invoices)
	if !payers["northwind labs"] {
		t.Error("missing northwind labs")
	}
	if !payers["brightloop"] {
		t.Error("missing brightloop")
	}
	if !payers["acme digital"] {
		t.Error("missing acme digital")
	}
	if payers["unknown"] {
		t.Error("unexpected unknown payer")
	}
}
