package tax

import (
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func sampleScale() Scale {
	return Scale{
		ValidFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Rows: []CategoryRow{
			{Category: CatA, CapARS: money.New(1200941045, money.ARS), ServicesFeeARS: money.New(4952718, money.ARS)},
			{Category: CatB, CapARS: money.New(1759518300, money.ARS), ServicesFeeARS: money.New(5637908, money.ARS)},
			{Category: CatK, CapARS: money.New(12661083875, money.ARS), ServicesFeeARS: money.New(161444604, money.ARS)},
		},
	}
}

func TestCategoryComparison(t *testing.T) {
	if !CatA.Less(CatB) {
		t.Fatal("A debe ser menor que B")
	}
	if CatK.Less(CatA) {
		t.Fatal("K no es menor que A")
	}
}

func TestCategoryFor(t *testing.T) {
	s := sampleScale()

	// Exactamente en el tope de A -> A
	cat, err := s.CategoryFor(money.New(1200941045, money.ARS))
	if err != nil || cat != CatA {
		t.Fatalf("esperaba A, got %v %v", cat, err)
	}

	// 1 centavo arriba del tope de A -> B
	cat, err = s.CategoryFor(money.New(1200941046, money.ARS))
	if err != nil || cat != CatB {
		t.Fatalf("esperaba B, got %v %v", cat, err)
	}

	// En el tope de K -> K
	cat, err = s.CategoryFor(money.New(12661083875, money.ARS))
	if err != nil || cat != CatK {
		t.Fatalf("esperaba K, got %v %v", cat, err)
	}

	// 1 centavo arriba de K -> ErrExceedsMaxCategory
	_, err = s.CategoryFor(money.New(12661083876, money.ARS))
	if !errors.Is(err, ErrExceedsMaxCategory) {
		t.Fatalf("esperaba ErrExceedsMaxCategory, got %v", err)
	}

	// Error si no es ARS
	_, err = s.CategoryFor(money.New(100, money.USD))
	if err == nil {
		t.Fatal("esperaba error por moneda no ARS")
	}
}

func TestScaleRowAndMaxCap(t *testing.T) {
	s := sampleScale()

	row, ok := s.Row(CatB)
	if !ok || row.CapARS.Amount != 1759518300 {
		t.Fatalf("Row B = %+v", row)
	}

	_, ok = s.Row("Z")
	if ok {
		t.Fatal("categoría Z no debería existir")
	}

	maxCap := s.MaxCap()
	if maxCap.Amount != 12661083875 {
		t.Fatalf("MaxCap = %v", maxCap)
	}

	emptyScale := Scale{}
	if !emptyScale.MaxCap().IsZero() {
		t.Fatal("MaxCap de escala vacía debe ser 0")
	}
}
