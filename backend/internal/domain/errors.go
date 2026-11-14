// Package domain contiene las entidades y reglas puras. Este archivo define los errores
// de dominio tipados que la capa HTTP mapea a status codes (architecture.md §2.3).
package domain

import (
	"errors"
	"sort"
	"strings"
)

// Errores centinela.
var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrInvalidTransition = errors.New("invalid transition")
	ErrForbidden         = errors.New("forbidden")
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrValidation        = errors.New("validation error")
	ErrVersionConflict   = errors.New("version conflict")
	ErrAlreadyApproved   = errors.New("already approved")
)

// ValidationError es un error de validación de un campo.
type ValidationError struct {
	Field string
	Msg   string
}

func (e ValidationError) Error() string { return e.Field + ": " + e.Msg }

// Unwrap permite errors.Is(err, ErrValidation).
func (e ValidationError) Unwrap() error { return ErrValidation }

// ValidationErrors agrupa errores de varios campos.
type ValidationErrors []ValidationError

func (es ValidationErrors) Error() string {
	parts := make([]string, len(es))
	for i, e := range es {
		parts[i] = e.Error()
	}
	return strings.Join(parts, "; ")
}

// Unwrap permite errors.Is(err, ErrValidation).
func (es ValidationErrors) Unwrap() error { return ErrValidation }

// Fields devuelve los errores como mapa campo → mensaje (orden estable al serializar).
func (es ValidationErrors) Fields() map[string]string {
	m := make(map[string]string, len(es))
	for _, e := range es {
		if _, ok := m[e.Field]; !ok {
			m[e.Field] = e.Msg
		}
	}
	return m
}

// OrNil devuelve nil si no hay errores.
func (es ValidationErrors) OrNil() error {
	if len(es) == 0 {
		return nil
	}
	sort.SliceStable(es, func(i, j int) bool { return es[i].Field < es[j].Field })
	return es
}

// Error es un error de dominio con un código de API específico (ej. VERSION_CONFLICT).
// Kind es uno de los centinelas y define el status HTTP.
type Error struct {
	Kind    error
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// Unwrap devuelve el centinela.
func (e *Error) Unwrap() error { return e.Kind }

// NewError crea un error con código.
func NewError(kind error, code, msg string) *Error {
	return &Error{Kind: kind, Code: code, Message: msg}
}
