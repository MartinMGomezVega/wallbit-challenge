// Package tax implementa las escalas de monotributo, categorías y liquidaciones impositivas
// (domain-context.md §2.1, §2.2, §7, architecture.md §2.2).
package tax

import (
	"errors"
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

// Category representa una categoría de Monotributo ("A".."K").
type Category string

const (
	CatA Category = "A"
	CatB Category = "B"
	CatC Category = "C"
	CatD Category = "D"
	CatE Category = "E"
	CatF Category = "F"
	CatG Category = "G"
	CatH Category = "H"
	CatI Category = "I"
	CatJ Category = "J"
	CatK Category = "K"
)

// Less reporta si c es menor que o en el orden A < B < ... < K.
func (c Category) Less(o Category) bool {
	return c < o
}

// CategoryRow es una fila de la escala de Monotributo para una categoría.
type CategoryRow struct {
	Category       Category    `json:"category"`
	CapARS         money.Money `json:"cap"`
	ServicesFeeARS money.Money `json:"servicesFee"`
}

// Scale es la tabla de categorías vigente desde ValidFrom.
type Scale struct {
	ValidFrom time.Time     `json:"validFrom"`
	Rows      []CategoryRow `json:"rows"` // Ordenadas A..K
}

// ErrExceedsMaxCategory se produce cuando la facturación excede el tope de la categoría máxima (K).
var ErrExceedsMaxCategory = errors.New("excede el tope máximo de Monotributo (categoría K)")

// ErrCategoryNotFound indica que la categoría solicitada no existe en la escala.
var ErrCategoryNotFound = errors.New("categoría no encontrada en la escala")

// CategoryFor devuelve la menor categoría con cap >= amount, o ErrExceedsMaxCategory si excede K.
func (s Scale) CategoryFor(amount money.Money) (Category, error) {
	if amount.Currency != money.ARS {
		return "", fmt.Errorf("la escala requiere ARS, got %s", amount.Currency)
	}
	for _, row := range s.Rows {
		if amount.Amount <= row.CapARS.Amount {
			return row.Category, nil
		}
	}
	return "", ErrExceedsMaxCategory
}

// Row devuelve los datos de la categoría c en la escala.
func (s Scale) Row(c Category) (CategoryRow, bool) {
	for _, r := range s.Rows {
		if r.Category == c {
			return r, true
		}
	}
	return CategoryRow{}, false
}

// MaxCap devuelve el tope anual de la categoría máxima (K).
func (s Scale) MaxCap() money.Money {
	if len(s.Rows) == 0 {
		return money.Zero(money.ARS)
	}
	return s.Rows[len(s.Rows)-1].CapARS
}
