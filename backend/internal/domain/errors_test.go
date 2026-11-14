package domain

import (
	"errors"
	"testing"
)

func TestValidationErrors(t *testing.T) {
	var es ValidationErrors
	if es.OrNil() != nil {
		t.Fatal("vacío debe ser nil")
	}
	es = append(es, ValidationError{"b", "mal"}, ValidationError{"a", "peor"}, ValidationError{"a", "dup"})
	err := es.OrNil()
	if !errors.Is(err, ErrValidation) {
		t.Fatal("no es ErrValidation")
	}
	if err.Error() != "a: peor; a: dup; b: mal" {
		t.Fatal(err.Error())
	}
	f := es.Fields()
	if f["a"] != "peor" || f["b"] != "mal" {
		t.Fatal(f)
	}
	single := ValidationError{"x", "y"}
	if !errors.Is(single, ErrValidation) || single.Error() != "x: y" {
		t.Fatal(single)
	}
}

func TestCodedError(t *testing.T) {
	e := NewError(ErrConflict, "VERSION_CONFLICT", "cambió")
	if !errors.Is(e, ErrConflict) || e.Error() != "VERSION_CONFLICT: cambió" {
		t.Fatal(e)
	}
	var target *Error
	if !errors.As(error(e), &target) || target.Code != "VERSION_CONFLICT" {
		t.Fatal("As")
	}
}
