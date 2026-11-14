package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
)

type overrideClassificationRequest struct {
	Kind string `json:"kind"`
	Note string `json:"note"`
}

// ClassificationDTO es la representación JSON de la clasificación de un movimiento.
type ClassificationDTO struct {
	MovementID   string `json:"movementId"`
	Kind         string `json:"kind"`
	Confidence   string `json:"confidence"`
	Rule         string `json:"rule"`
	OverriddenBy string `json:"overriddenBy,omitempty"`
	OverriddenAt string `json:"overriddenAt,omitempty"`
	Note         string `json:"note,omitempty"`
}

func toClassificationDTO(c movement.Classification) ClassificationDTO {
	var overriddenAt string
	if c.OverriddenAt != nil && !c.OverriddenAt.IsZero() {
		overriddenAt = c.OverriddenAt.Format(time.RFC3339)
	}
	return ClassificationDTO{
		MovementID:   c.MovementID,
		Kind:         string(c.Kind),
		Confidence:   string(c.Confidence),
		Rule:         string(c.Rule),
		OverriddenBy: c.OverriddenBy,
		OverriddenAt: overriddenAt,
		Note:         c.Note,
	}
}

type movementsHandler struct {
	overrideUsecase *usecase.OverrideClassification
}

func (h *movementsHandler) overrideClassification(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ID de movimiento requerido", nil)
		return
	}

	var req overrideClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "JSON inválido", nil)
		return
	}

	operator := "accountant"
	if ident, ok := FromContext(r.Context()); ok && ident.Role != "" {
		operator = ident.Role
	}

	res, err := h.overrideUsecase.Execute(r.Context(), usecase.OverrideClassificationInput{
		MovementID: id,
		Kind:       movement.ClassificationKind(req.Kind),
		Note:       req.Note,
		Operator:   operator,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toClassificationDTO(res))
}
