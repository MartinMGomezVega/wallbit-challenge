package signal

import (
	"testing"
)

func TestCatalog(t *testing.T) {
	items := Catalog()
	if len(items) != 13 {
		t.Fatalf("expected 13 catalog items, got %d", len(items))
	}

	// Verificar orden de prioridad estrictamente ascendente (1 a 13)
	for i, item := range items {
		expectedPrio := i + 1
		if item.Priority != expectedPrio {
			t.Errorf("item[%d] priority = %d, want %d", i, item.Priority, expectedPrio)
		}
		if item.Code == "" {
			t.Errorf("item[%d] code is empty", i)
		}
		if item.Label == "" {
			t.Errorf("item[%d] label is empty", i)
		}
		if item.Severity != "RED" && item.Severity != "YELLOW" {
			t.Errorf("item[%d] unexpected severity: %s", i, item.Severity)
		}
	}

	// MetaFor
	meta, ok := MetaFor(CodeMonoExclusionKCapExceeded)
	if !ok || meta.Priority != 1 || meta.Severity != SeverityRed {
		t.Errorf("MetaFor(CodeMonoExclusionKCapExceeded) = %+v, %v", meta, ok)
	}

	_, okNone := MetaFor("NON_EXISTENT")
	if okNone {
		t.Error("MetaFor(NON_EXISTENT) should return false")
	}

	// Severity.String()
	if SeverityGreen.String() != "GREEN" {
		t.Errorf("SeverityGreen.String() = %s", SeverityGreen.String())
	}
	if Severity(99).String() != "UNKNOWN" {
		t.Errorf("Severity(99).String() = %s", Severity(99).String())
	}
}
