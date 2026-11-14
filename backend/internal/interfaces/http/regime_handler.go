package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

type regimeHandler struct {
	changeRegimeUC    *usecase.ChangeRegime
	updateRIProfileUC *usecase.UpdateRIProfile
}

type regimeChangeRequest struct {
	To            string `json:"to"`
	EffectiveDate string `json:"effectiveDate"`
	Note          string `json:"note"`
	Force         bool   `json:"force"`
}

type riProfileRequest struct {
	SelfEmployedMonthlyContribution money.Money           `json:"selfEmployedMonthlyContribution"`
	PreviousVatCredit               money.Money           `json:"previousVatCredit"`
	IncomeTaxAdvances               []incomeTaxAdvanceDTO `json:"incomeTaxAdvances"`
}

type incomeTaxAdvanceDTO struct {
	Period  string      `json:"period"`
	Amount  money.Money `json:"amount"`
	DueDate time.Time   `json:"dueDate"`
}

func (h *regimeHandler) changeRegime(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("id")
	if clientID == "" {
		writeErrorCode(w, http.StatusNotFound, "NOT_FOUND", "ID de cliente no provisto", nil)
		return
	}

	var req regimeChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "JSON inválido", nil)
		return
	}

	if req.To != "RI" {
		writeErrorCode(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "el campo 'to' debe ser 'RI'", nil)
		return
	}

	effDate, err := time.ParseInLocation("2006-01-02", req.EffectiveDate, period.Location())
	if err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "fecha efectiva inválida (esperado YYYY-MM-DD)", nil)
		return
	}

	c, err := h.changeRegimeUC.Execute(r.Context(), usecase.ChangeRegimeInput{
		ClientID:      clientID,
		To:            client.RegimeRI,
		EffectiveDate: effDate,
		Note:          req.Note,
		Force:         req.Force,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *regimeHandler) updateRIProfile(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("id")
	if clientID == "" {
		writeErrorCode(w, http.StatusNotFound, "NOT_FOUND", "ID de cliente no provisto", nil)
		return
	}

	var req riProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "JSON inválido", nil)
		return
	}

	advances := make([]tax.IncomeTaxAdvance, 0, len(req.IncomeTaxAdvances))
	for _, advDTO := range req.IncomeTaxAdvances {
		p, err := period.Parse(advDTO.Period)
		if err != nil {
			writeErrorCode(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "período de anticipo inválido", nil)
			return
		}
		advances = append(advances, tax.IncomeTaxAdvance{
			Period:  p,
			Amount:  advDTO.Amount,
			DueDate: advDTO.DueDate,
		})
	}

	prof, err := h.updateRIProfileUC.Execute(r.Context(), usecase.UpdateRIProfileInput{
		ClientID:                        clientID,
		SelfEmployedMonthlyContribution: req.SelfEmployedMonthlyContribution,
		PreviousVatCredit:               req.PreviousVatCredit,
		IncomeTaxAdvances:               advances,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, prof)
}
