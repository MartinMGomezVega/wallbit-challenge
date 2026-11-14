package mock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestScalesExactValues(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() falló: %v", err)
	}

	ctx := context.Background()

	// 1. Escala agosto 2026 (evaluada al 2026-10-15)
	sAug, err := c.MonotributoScale(ctx, time.Date(2026, 10, 15, 0, 0, 0, 0, period.Location()))
	if err != nil {
		t.Fatal(err)
	}

	// Verificar A, H, K
	rowA, _ := sAug.Row(tax.CatA)
	if rowA.CapARS.Amount != 1200941045 || rowA.ServicesFeeARS.Amount != 4952718 {
		t.Fatalf("Aug Cat A = %+v", rowA)
	}
	rowH, _ := sAug.Row(tax.CatH)
	if rowH.CapARS.Amount != 8192466037 || rowH.ServicesFeeARS.Amount != 52270668 {
		t.Fatalf("Aug Cat H = %+v", rowH)
	}
	rowK, _ := sAug.Row(tax.CatK)
	if rowK.CapARS.Amount != 12661083875 || rowK.ServicesFeeARS.Amount != 161444604 {
		t.Fatalf("Aug Cat K = %+v", rowK)
	}
	if sAug.MaxCap().Amount != 12661083875 {
		t.Fatalf("MaxCap = %v", sAug.MaxCap())
	}

	// 2. Escala febrero 2026 (evaluada al 2026-03-01)
	sFeb, err := c.MonotributoScale(ctx, time.Date(2026, 3, 1, 0, 0, 0, 0, period.Location()))
	if err != nil {
		t.Fatal(err)
	}
	rowFebA, _ := sFeb.Row(tax.CatA)
	if rowFebA.CapARS.Amount != 1027798813 || rowFebA.ServicesFeeARS.Amount != 4238674 {
		t.Fatalf("Feb Cat A = %+v", rowFebA)
	}
	rowFebH, _ := sFeb.Row(tax.CatH)
	if rowFebH.CapARS.Amount != 7011340733 || rowFebH.ServicesFeeARS.Amount != 44734693 {
		t.Fatalf("Feb Cat H = %+v", rowFebH)
	}
	rowFebK, _ := sFeb.Row(tax.CatK)
	if rowFebK.CapARS.Amount != 10835708400 || rowFebK.ServicesFeeARS.Amount != 138168790 {
		t.Fatalf("Feb Cat K = %+v", rowFebK)
	}

	// 3. Fecha anterior a febrero 2026 (2025-12-01) -> escala de febrero (supuesto §2.2 nota)
	sDec, err := c.MonotributoScale(ctx, time.Date(2025, 12, 1, 0, 0, 0, 0, period.Location()))
	if err != nil {
		t.Fatal(err)
	}
	if sDec.MaxCap().Amount != 10835708400 {
		t.Fatalf("Dec 2025 debe devolver escala feb, got maxCap=%v", sDec.MaxCap())
	}
}

func TestDueDatesRealAndReferential(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// 1. Caso noviembre 2026 (real): CUIT terminación 8
	pNov, _ := period.Parse("2026-11")
	cuit := client.CUIT("20-12345678-9") // terminación 8

	dd, err := c.DueDates(ctx, pNov, cuit)
	if err != nil {
		t.Fatal(err)
	}
	if dd.Monotributo.Format("2006-01-02") != "2026-11-20" {
		t.Fatalf("Monotributo = %s, want 2026-11-20", dd.Monotributo.Format("2006-01-02"))
	}
	if dd.VAT.Format("2006-01-02") != "2026-11-25" {
		t.Fatalf("VAT = %s, want 2026-11-25", dd.VAT.Format("2006-01-02"))
	}
	if dd.SelfEmployed.Format("2006-01-02") != "2026-11-11" {
		t.Fatalf("SelfEmployed = %s, want 2026-11-11", dd.SelfEmployed.Format("2006-01-02"))
	}

	// 2. Patrón referencial en un mes donde día 20 cae sábado:
	// Junio 2026: 20/06/2026 es sábado -> Monotributo pasa al lunes 22
	pJun, _ := period.Parse("2026-06")
	ddJun, err := c.DueDates(ctx, pJun, client.CUIT("20-35000000-4")) // terminación 0
	if err != nil {
		t.Fatal(err)
	}
	if ddJun.Monotributo.Format("2006-01-02") != "2026-06-22" {
		t.Fatalf("Jun Monotributo en sábado debe pasar a lunes: %s", ddJun.Monotributo.Format("2006-01-02"))
	}
	// IVA terminación 0 -> día 18 (jueves 18)
	if ddJun.VAT.Format("2006-01-02") != "2026-06-18" {
		t.Fatalf("Jun VAT = %s", ddJun.VAT.Format("2006-01-02"))
	}
	// Autónomos terminación 0 -> día 5 (viernes 5)
	if ddJun.SelfEmployed.Format("2006-01-02") != "2026-06-05" {
		t.Fatalf("Jun Autónomos = %s", ddJun.SelfEmployed.Format("2006-01-02"))
	}
}

func TestClientOptionsAndFailures(t *testing.T) {
	// FailEvery
	c, err := New(WithFailEvery(2))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	p, _ := period.Parse("2026-11")
	cuit := client.CUIT("20-12345678-9")

	_, err = c.DueDates(ctx, p, cuit) // llamada 1: ok
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.DueDates(ctx, p, cuit) // llamada 2: falla
	if !errors.Is(err, ErrSimulatedFailure) {
		t.Fatalf("esperaba ErrSimulatedFailure, got %v", err)
	}

	// Context cancelado con Latency
	cLat, _ := New(WithLatency(50 * time.Millisecond))
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = cLat.MonotributoScale(cancelCtx, time.Now())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("esperaba context.Canceled, got %v", err)
	}
}
