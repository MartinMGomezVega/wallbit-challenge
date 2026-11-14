package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
)

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// writeJSON serializa v como JSON con el status dado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErrorCode escribe un error con código y status explícitos.
func writeErrorCode(w http.ResponseWriter, status int, code, msg string, fields map[string]string) {
	writeJSON(w, status, errorBody{Error: errorPayload{Code: code, Message: msg, Fields: fields}})
}

// writeError mapea un error de dominio a la respuesta HTTP (architecture.md §2.3).
func writeError(w http.ResponseWriter, err error) {
	status, payload := mapError(err)
	if status == http.StatusInternalServerError {
		slog.Error("internal error", "error", err.Error())
	}
	writeJSON(w, status, errorBody{Error: payload})
}

func mapError(err error) (int, errorPayload) {
	var ves domain.ValidationErrors
	if errors.As(err, &ves) {
		return http.StatusUnprocessableEntity, errorPayload{Code: "VALIDATION_ERROR", Message: "Datos inválidos", Fields: ves.Fields()}
	}
	var ve domain.ValidationError
	if errors.As(err, &ve) {
		return http.StatusUnprocessableEntity, errorPayload{Code: "VALIDATION_ERROR", Message: "Datos inválidos", Fields: map[string]string{ve.Field: ve.Msg}}
	}
	status, code := statusFor(err)
	var ce *domain.Error
	if errors.As(err, &ce) {
		return status, errorPayload{Code: ce.Code, Message: ce.Message}
	}
	msg := err.Error()
	if status == http.StatusInternalServerError {
		msg = "Error interno"
	}
	return status, errorPayload{Code: code, Message: msg}
}

func statusFor(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity, "VALIDATION_ERROR"
	case errors.Is(err, domain.ErrVersionConflict):
		return http.StatusConflict, "VERSION_CONFLICT"
	case errors.Is(err, domain.ErrAlreadyApproved):
		return http.StatusConflict, "ALREADY_APPROVED"
	case errors.Is(err, domain.ErrInvalidTransition):
		return http.StatusConflict, "INVALID_TRANSITION"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}
