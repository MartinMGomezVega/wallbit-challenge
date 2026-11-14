package invoice

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func TestParseMisComprobantesOK(t *testing.T) {
	path := filepath.Join("testdata", "mis_comprobantes_ok.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	rows, warnings, lineErrors, err := ParseMisComprobantes(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseMisComprobantes unexpected error: %v", err)
	}

	if len(lineErrors) != 0 {
		t.Fatalf("unexpected line errors: %+v", lineErrors)
	}

	// 4 facturas válidas (3 E y 1 C)
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}

	// 1 nota de crédito como advertencia
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for credit note, got %d", len(warnings))
	}
	if warnings[0].Line != 5 {
		t.Errorf("warning line = %d, want 5", warnings[0].Line)
	}

	// Verificar fila 1: Factura E de USD 1500
	r0 := rows[0]
	if r0.Type != TypeE || r0.PointOfSale != 1 || r0.Number != 1 {
		t.Errorf("r0 type/pos/num mismatch: %+v", r0)
	}
	if r0.Currency != money.USD || r0.TotalAmount.Amount != 150000 {
		t.Errorf("r0 amount mismatch: %+v", r0.TotalAmount)
	}
	if r0.FXRate == nil || *r0.FXRate != money.Rate(14871500) {
		t.Errorf("r0 FXRate mismatch: %v", r0.FXRate)
	}

	// Verificar fila 3: Factura C de ARS 500,000
	r2 := rows[2]
	if r2.Type != TypeC || r2.Currency != money.ARS || r2.TotalAmount.Amount != 50000000 {
		t.Errorf("r2 mismatch: %+v", r2)
	}
}

func TestParseMisComprobantesSemicolon(t *testing.T) {
	path := filepath.Join("testdata", "mis_comprobantes_semicolon.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	rows, warnings, lineErrors, err := ParseMisComprobantes(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseMisComprobantes unexpected error: %v", err)
	}
	if len(lineErrors) != 0 {
		t.Fatalf("unexpected line errors: %+v", lineErrors)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}

	// 3.000,00 USD
	if rows[0].TotalAmount.Amount != 300000 {
		t.Errorf("row 0 amount = %v, want 300000", rows[0].TotalAmount.Amount)
	}
}

func TestParseMisComprobantesBadRows(t *testing.T) {
	path := filepath.Join("testdata", "mis_comprobantes_bad_rows.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	rows, _, lineErrors, err := ParseMisComprobantes(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseMisComprobantes unexpected error: %v", err)
	}

	// Solo la última fila (línea 8) es válida
	if len(rows) != 1 {
		t.Fatalf("expected 1 valid row, got %d", len(rows))
	}
	if rows[0].Number != 6 {
		t.Errorf("valid row number = %d, want 6", rows[0].Number)
	}

	if len(lineErrors) != 6 {
		t.Fatalf("expected 6 line errors, got %d: %+v", len(lineErrors), lineErrors)
	}
}

func TestParseMisComprobantesErrors(t *testing.T) {
	// 1. Archivo vacío
	_, _, _, err := ParseMisComprobantes(bytes.NewReader([]byte("")))
	if err == nil {
		t.Fatal("expected error on empty reader")
	}

	// 2. Encabezados faltantes
	badHeaders := "Fecha,Tipo,Moneda\n01/10/2026,E,USD"
	_, _, _, err = ParseMisComprobantes(bytes.NewReader([]byte(badHeaders)))
	if err == nil {
		t.Fatal("expected error on missing headers")
	}
}

func TestParseMisComprobantesDemoPaula(t *testing.T) {
	path := filepath.Join("testdata", "mis_comprobantes_demo_paula.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	rows, warnings, lineErrors, err := ParseMisComprobantes(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseMisComprobantes error: %v", err)
	}
	if len(lineErrors) != 0 || len(warnings) != 0 {
		t.Fatalf("unexpected line errors or warnings: %v %v", lineErrors, warnings)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].TotalAmount.Amount != 85000 || rows[0].ReceiverName != "Brightloop GmbH" {
		t.Fatalf("unexpected row data: %+v", rows[0])
	}
}

