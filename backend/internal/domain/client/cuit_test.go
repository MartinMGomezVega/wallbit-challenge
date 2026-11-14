package client

import (
	"errors"
	"testing"
)

func TestParseCUIT(t *testing.T) {
	valid := []string{
		"20-35000000-4",
		"20-35000001-2",
		"20-35000002-0",
		"20-35000003-9",
		"20-35000004-7",
		"20-35000005-5",
		"20-35000006-3",
		"20-35000007-1",
		"20-35000009-8",
		"20-12345678-6",
	}

	for _, s := range valid {
		cuit, err := ParseCUIT(s)
		if err != nil {
			t.Fatalf("ParseCUIT(%s) falló: %v", s, err)
		}
		if cuit.String() != s {
			t.Fatalf("String = %s, want %s", cuit.String(), s)
		}
	}

	// Invalid format
	invalidFormat := []string{
		"",
		"20350000004",
		"20-35000000",
		"20-35000000-44",
		"abc",
		"20-1234abcd-1",
	}
	for _, s := range invalidFormat {
		if _, err := ParseCUIT(s); !errors.Is(err, ErrInvalidCUITFormat) {
			t.Fatalf("esperaba ErrInvalidCUITFormat para %q, got %v", s, err)
		}
	}

	// Invalid checksum
	invalidChecksum := []string{
		"20-35000000-5", // debería ser 4
		"20-12345678-9", // debería ser 6
	}
	for _, s := range invalidChecksum {
		if _, err := ParseCUIT(s); !errors.Is(err, ErrInvalidCUITChecksum) {
			t.Fatalf("esperaba ErrInvalidCUITChecksum para %q, got %v", s, err)
		}
	}
}

func TestTerminacionAndMasked(t *testing.T) {
	// Terminación = penúltimo dígito antes del verificador
	// 20-12345678-9 -> terminación 8
	cuit := CUIT("20-12345678-9")
	if cuit.Terminacion() != 8 {
		t.Fatalf("Terminacion = %d, want 8", cuit.Terminacion())
	}

	cuit2 := CUIT("27-35000004-7")
	if cuit2.Terminacion() != 4 {
		t.Fatalf("Terminacion = %d, want 4", cuit2.Terminacion())
	}

	cuitZero := CUIT("20-35000000-4")
	if cuitZero.Terminacion() != 0 {
		t.Fatalf("Terminacion = %d, want 0", cuitZero.Terminacion())
	}

	// Masked: "20-****5678-9"
	if cuit.Masked() != "20-****5678-9" {
		t.Fatalf("Masked = %s, want 20-****5678-9", cuit.Masked())
	}
	if CUIT("corto").Masked() != "corto" {
		t.Fatal("Masked con longitud distinta debe devolver s")
	}
	if CUIT("").Terminacion() != 0 {
		t.Fatal("Terminacion con CUIT corto debe ser 0")
	}
}
