package usecase

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestInvoicesUsecases(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 2, 12, 0, 0, 0, period.Location())
	clk := fakes.Clock{Current: now}
	idGen := fakes.NewIDGenerator()

	clientRepo := fakes.NewClientRepository()
	invRepo := fakes.NewInvoiceRepository()
	piRepo := fakes.NewPurchaseInvoiceRepository()
	ciRepo := fakes.NewClientInputRepository()
	notifier := &recordingNotifier{}

	cMono := client.Client{
		ID:     "cli_mono",
		Name:   "Monotributista Test",
		Regime: client.RegimeMonotributo,
	}
	cRI := client.Client{
		ID:     "cli_ri",
		Name:   "RI Test",
		Regime: client.RegimeRI,
	}
	_ = clientRepo.Save(ctx, cMono)
	_ = clientRepo.Save(ctx, cRI)

	importUC := &ImportInvoicesCSV{
		Clients:  clientRepo,
		Invoices: invRepo,
		IDGen:    idGen,
		Clock:    clk,
		Notifier: notifier,
	}

	addInvUC := &AddInvoice{
		Clients:  clientRepo,
		Invoices: invRepo,
		IDGen:    idGen,
		Clock:    clk,
		Notifier: notifier,
	}

	listInvUC := &ListInvoices{Invoices: invRepo}

	addPIUC := &AddPurchaseInvoice{
		Clients:          clientRepo,
		PurchaseInvoices: piRepo,
		IDGen:            idGen,
		Clock:            clk,
		Notifier:         notifier,
	}

	setIncomeUC := &SetExternalIncome{
		Inputs:   ciRepo,
		Clock:    clk,
		Notifier: notifier,
	}

	confirmUC := &ConfirmPeriodInput{
		Inputs:   ciRepo,
		Clock:    clk,
		Notifier: notifier,
	}

	t.Run("ImportInvoicesCSV: Success & Idempotency", func(t *testing.T) {
		csvContent := `Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total
05/10/2026,19 - Factura E,00001,00000001,00000001,74392819283746,80,30712345678,Northwind Labs LLC,1487.15,USD,1500.00,0.00,0.00,0.00,1500.00
15/10/2026,11 - Factura C,00001,00000002,00000002,74392819283748,80,30712345680,Cliente Local SA,1.00,$,500000.00,0.00,0.00,0.00,500000.00
`
		res1, err := importUC.Execute(ctx, "cli_mono", strings.NewReader(csvContent))
		if err != nil {
			t.Fatalf("first import failed: %v", err)
		}
		if res1.Imported != 2 || res1.SkippedDuplicates != 0 {
			t.Fatalf("first import expected 2 imported, got %+v", res1)
		}

		// Idempotency: second import of same CSV
		res2, err := importUC.Execute(ctx, "cli_mono", strings.NewReader(csvContent))
		if err != nil {
			t.Fatalf("second import failed: %v", err)
		}
		if res2.Imported != 0 || res2.SkippedDuplicates != 2 {
			t.Fatalf("second import expected 0 imported, 2 skipped, got %+v", res2)
		}
	})

	t.Run("ImportInvoicesCSV: Regime mismatch", func(t *testing.T) {
		// Monotributista importing Factura A
		csvContent := `Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total
05/10/2026,1 - Factura A,00001,00000010,00000010,74392819283746,80,30712345678,Cliente Local SA,1.00,$,100000.00,0.00,0.00,21000.00,121000.00
`
		res, err := importUC.Execute(ctx, "cli_mono", strings.NewReader(csvContent))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Imported != 0 || len(res.Errors) != 1 {
			t.Fatalf("expected 0 imported and 1 error for Factura A in Monotributo, got %+v", res)
		}
	})

	t.Run("AddInvoice: Success, validation & conflict", func(t *testing.T) {
		inv := invoice.Invoice{
			ClientID:     "cli_mono",
			Type:         invoice.TypeE,
			PointOfSale:  2,
			Number:       5,
			IssueDate:    time.Date(2026, 10, 20, 10, 0, 0, 0, period.Location()),
			Currency:     money.USD,
			TotalAmount:  money.New(80000, money.USD),
			ReceiverName: "Foreign Client Ltd",
		}

		created, err := addInvUC.Execute(ctx, inv)
		if err != nil {
			t.Fatalf("AddInvoice failed: %v", err)
		}
		if created.ID == "" || created.Source != invoice.SourceManual {
			t.Fatalf("unexpected created invoice: %+v", created)
		}

		// Conflict: duplicate
		_, err = addInvUC.Execute(ctx, inv)
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict on duplicate, got %v", err)
		}
	})

	t.Run("ListInvoices: With and without period", func(t *testing.T) {
		p, _ := period.Parse("2026-10")
		listWithPeriod, err := listInvUC.Execute(ctx, "cli_mono", &p)
		if err != nil || len(listWithPeriod) == 0 {
			t.Fatalf("ListInvoices with period failed: %v, count %d", err, len(listWithPeriod))
		}

		listAll, err := listInvUC.Execute(ctx, "cli_mono", nil)
		if err != nil || len(listAll) == 0 {
			t.Fatalf("ListInvoices all failed: %v", err)
		}
	})

	t.Run("AddPurchaseInvoice: Success for RI, forbidden for Monotributo", func(t *testing.T) {
		piInput := AddPurchaseInvoiceInput{
			ClientID:     "cli_mono", // Monotributo -> should fail!
			IssueDate:    time.Date(2026, 10, 10, 0, 0, 0, 0, period.Location()),
			SupplierName: "Proveedor SRL",
			SupplierCUIT: "20-35000000-4",
			NetAmount:    money.New(1000000, money.ARS),
			VatAmount:    money.New(210000, money.ARS),
		}

		_, err := addPIUC.Execute(ctx, piInput)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("expected ErrForbidden for Monotributo purchase invoice, got %v", err)
		}

		// For RI client -> success
		piInput.ClientID = "cli_ri"
		createdPI, err := addPIUC.Execute(ctx, piInput)
		if err != nil {
			t.Fatalf("AddPurchaseInvoice for RI failed: %v", err)
		}
		if createdPI.ID == "" || createdPI.SupplierName != "Proveedor SRL" {
			t.Fatalf("unexpected created purchase invoice: %+v", createdPI)
		}
	})

	t.Run("External income & Confirm flow", func(t *testing.T) {
		p, _ := period.Parse("2026-10")

		// 1. Confirm before answering -> error
		_, err := confirmUC.Execute(ctx, "cli_ri", p)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation when confirming unanswered, got %v", err)
		}

		// 2. Set external income (none = true)
		resIncome, err := setIncomeUC.Execute(ctx, SetExternalIncomeInput{
			ClientID: "cli_ri",
			Period:   p,
			None:     true,
		})
		if err != nil || !resIncome.DeclaredNoExternalIncome {
			t.Fatalf("SetExternalIncome failed: %v, %+v", err, resIncome)
		}

		// 3. Confirm now succeeds
		resConfirm, err := confirmUC.Execute(ctx, "cli_ri", p)
		if err != nil || resConfirm.ConfirmedAt == nil {
			t.Fatalf("ConfirmPeriodInput failed: %v, %+v", err, resConfirm)
		}
	})

	t.Run("Invalid CSV header", func(t *testing.T) {
		_, err := importUC.Execute(ctx, "cli_mono", bytes.NewReader([]byte("bad headers\n1,2,3")))
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation on bad headers, got %v", err)
		}
	})
}
