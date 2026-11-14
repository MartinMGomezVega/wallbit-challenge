package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
)

// MoneyDTO representa un importe monetario para la API JSON.
type MoneyDTO struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// ParametersDTO es el formato de transferencia JSON para los parámetros del contador.
type ParametersDTO struct {
	InvoiceECountsForCap           bool     `json:"invoiceECountsForCap"`
	FXRule                         string   `json:"fxRule"`
	CategoryCapWarningBp           int64    `json:"categoryCapWarningBp"`
	KCapWarningBp                  int64    `json:"kCapWarningBp"`
	AtypicalMultiplierBp           int64    `json:"atypicalMultiplierBp"`
	AtypicalWindowMonths           int      `json:"atypicalWindowMonths"`
	MissingInvoiceMinUSD           MoneyDTO `json:"missingInvoiceMinUsd"`
	ReconciliationToleranceBp      int64    `json:"reconciliationToleranceBp"`
	ReconciliationWindowBeforeDays int      `json:"reconciliationWindowBeforeDays"`
	ReconciliationWindowAfterDays  int      `json:"reconciliationWindowAfterDays"`
	CardSpendKCapBp                int64    `json:"cardSpendKCapBp"`
	ReminderDaysBeforeDue          int      `json:"reminderDaysBeforeDue"`
	UpdatedAt                      string   `json:"updatedAt,omitempty"`
	UpdatedBy                      string   `json:"updatedBy,omitempty"`
}

type putParametersResponse struct {
	Parameters   ParametersDTO `json:"parameters"`
	Recalculated int           `json:"recalculated"`
}

func toParametersDTO(p parameters.AccountantParameters) ParametersDTO {
	var updatedAt string
	if !p.UpdatedAt.IsZero() {
		updatedAt = p.UpdatedAt.Format(time.RFC3339)
	}
	return ParametersDTO{
		InvoiceECountsForCap:           p.InvoiceECountsForCap,
		FXRule:                         string(p.FXRule),
		CategoryCapWarningBp:           int64(p.CategoryCapWarningBp),
		KCapWarningBp:                  int64(p.KCapWarningBp),
		AtypicalMultiplierBp:           int64(p.AtypicalMultiplierBp),
		AtypicalWindowMonths:           p.AtypicalWindowMonths,
		MissingInvoiceMinUSD:           MoneyDTO{Amount: p.MissingInvoiceMinUSD.DecimalString(), Currency: string(p.MissingInvoiceMinUSD.Currency)},
		ReconciliationToleranceBp:      int64(p.ReconciliationToleranceBp),
		ReconciliationWindowBeforeDays: p.ReconciliationWindowBeforeDays,
		ReconciliationWindowAfterDays:  p.ReconciliationWindowAfterDays,
		CardSpendKCapBp:                int64(p.CardSpendKCapBp),
		ReminderDaysBeforeDue:          p.ReminderDaysBeforeDue,
		UpdatedAt:                      updatedAt,
		UpdatedBy:                      p.UpdatedBy,
	}
}

func fromParametersDTO(dto ParametersDTO) (parameters.AccountantParameters, error) {
	minUSD, err := money.ParseDecimal(dto.MissingInvoiceMinUSD.Amount, money.Currency(dto.MissingInvoiceMinUSD.Currency))
	if err != nil {
		minUSD = money.New(-1, money.USD) // Forzará error de validación en Validate()
	}
	return parameters.AccountantParameters{
		InvoiceECountsForCap:           dto.InvoiceECountsForCap,
		FXRule:                         parameters.FXRule(dto.FXRule),
		CategoryCapWarningBp:           money.BasisPoints(dto.CategoryCapWarningBp),
		KCapWarningBp:                  money.BasisPoints(dto.KCapWarningBp),
		AtypicalMultiplierBp:           money.BasisPoints(dto.AtypicalMultiplierBp),
		AtypicalWindowMonths:           dto.AtypicalWindowMonths,
		MissingInvoiceMinUSD:           minUSD,
		ReconciliationToleranceBp:      money.BasisPoints(dto.ReconciliationToleranceBp),
		ReconciliationWindowBeforeDays: dto.ReconciliationWindowBeforeDays,
		ReconciliationWindowAfterDays:  dto.ReconciliationWindowAfterDays,
		CardSpendKCapBp:                money.BasisPoints(dto.CardSpendKCapBp),
		ReminderDaysBeforeDue:          dto.ReminderDaysBeforeDue,
	}, nil
}

type parametersHandler struct {
	getUsecase    *usecase.GetParameters
	updateUsecase *usecase.UpdateParameters
}

func (h *parametersHandler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.getUsecase.Execute(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toParametersDTO(p))
}

func (h *parametersHandler) update(w http.ResponseWriter, r *http.Request) {
	var dto ParametersDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "JSON inválido", nil)
		return
	}
	domainParams, _ := fromParametersDTO(dto)
	result, err := h.updateUsecase.Execute(r.Context(), domainParams)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, putParametersResponse{
		Parameters:   toParametersDTO(result.Parameters),
		Recalculated: result.Recalculated,
	})
}
