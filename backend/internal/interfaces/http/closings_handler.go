package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
)

type closingsHandler struct {
	runClosingUC      *usecase.RunMonthlyClosing
	listClosingsUC    *usecase.ListClosings
	listClosingIDsUC  *usecase.ListClosingIDs
	closingsSummaryUC *usecase.ClosingsSummary
	getDetailUC       *usecase.GetClosingDetail
	approveUC         *usecase.ApproveClosing
	approveBatchUC    *usecase.ApproveBatch
	sendRemindersUC   *usecase.SendPaymentReminders
	clock             ports.Clock
}

// parseClosingFilter parsea los filtros comunes desde la query string (reutilizable por spec 009).
func parseClosingFilter(r *http.Request) (closing.Filter, closing.Sort, int, int) {
	q := r.URL.Query()

	var filter closing.Filter

	// 1. Severities (CSV de RED, YELLOW, GREEN)
	if sevParam := q.Get("severity"); sevParam != "" {
		parts := strings.Split(sevParam, ",")
		for _, p := range parts {
			trimmed := strings.ToUpper(strings.TrimSpace(p))
			switch trimmed {
			case "RED":
				filter.Severities = append(filter.Severities, signal.SeverityRed)
			case "YELLOW":
				filter.Severities = append(filter.Severities, signal.SeverityYellow)
			case "GREEN":
				filter.Severities = append(filter.Severities, signal.SeverityGreen)
			}
		}
	}

	// 2. Signals (CSV de códigos)
	if sigParam := q.Get("signal"); sigParam != "" {
		parts := strings.Split(sigParam, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				filter.Signals = append(filter.Signals, signal.Code(trimmed))
			}
		}
	}

	// 3. Regime
	if regParam := q.Get("regime"); regParam != "" {
		trimmed := strings.ToUpper(strings.TrimSpace(regParam))
		if trimmed == "MONOTRIBUTO" || trimmed == "RI" {
			filter.Regime = client.Regime(trimmed)
		}
	}

	// 4. Status (CSV)
	if stParam := q.Get("status"); stParam != "" {
		parts := strings.Split(stParam, ",")
		for _, p := range parts {
			trimmed := strings.ToUpper(strings.TrimSpace(p))
			if trimmed != "" {
				filter.Statuses = append(filter.Statuses, closing.Status(trimmed))
			}
		}
	}

	// 5. Query
	filter.Query = q.Get("q")

	// Sort
	sortCol := q.Get("sort")
	if sortCol == "" {
		sortCol = "severity"
	}
	sortDir := q.Get("dir")
	s := closing.Sort{
		By:  sortCol,
		Dir: sortDir,
	}

	// Pagination
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}

	return filter, s, page, pageSize
}

// run maneja POST /api/v1/accountant/periods/{period}/run
func (h *closingsHandler) run(w http.ResponseWriter, r *http.Request) {
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "INVALID_PERIOD", "Formato de período inválido (esperado YYYY-MM)", nil)
		return
	}

	summary, err := h.runClosingUC.Execute(r.Context(), p)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// list maneja GET /api/v1/accountant/periods/{period}/closings
func (h *closingsHandler) list(w http.ResponseWriter, r *http.Request) {
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "INVALID_PERIOD", "Formato de período inválido (esperado YYYY-MM)", nil)
		return
	}

	filter, s, page, pageSize := parseClosingFilter(r)

	res, err := h.listClosingsUC.Execute(r.Context(), usecase.ListClosingsQuery{
		Period:   p,
		Filter:   filter,
		Sort:     s,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// ids maneja GET /api/v1/accountant/periods/{period}/closings/ids (spec 009 criterio 1).
func (h *closingsHandler) ids(w http.ResponseWriter, r *http.Request) {
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "INVALID_PERIOD", "Formato de período inválido (esperado YYYY-MM)", nil)
		return
	}

	filter, s, _, _ := parseClosingFilter(r)

	res, err := h.listClosingIDsUC.Execute(r.Context(), usecase.ListClosingIDsQuery{
		Period: p,
		Filter: filter,
		Sort:   s,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// summary maneja GET /api/v1/accountant/periods/{period}/closings/summary
func (h *closingsHandler) summary(w http.ResponseWriter, r *http.Request) {
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "INVALID_PERIOD", "Formato de período inválido (esperado YYYY-MM)", nil)
		return
	}

	filter, _, _, _ := parseClosingFilter(r)

	sum, err := h.closingsSummaryUC.Execute(r.Context(), p, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sum)
}

// detail maneja GET /api/v1/accountant/closings/{id}
func (h *closingsHandler) detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ID de cierre requerido", nil)
		return
	}

	det, err := h.getDetailUC.Execute(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, det)
}

type approveRequest struct {
	Version    int    `json:"version"`
	ClientNote string `json:"clientNote"`
}

type approveResponse struct {
	closing.Closing
	Warning string `json:"warning,omitempty"`
}

// approve maneja POST /api/v1/accountant/closings/{id}/approve (spec 010 criterio 1 y 2).
func (h *closingsHandler) approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ID de cierre requerido", nil)
		return
	}

	var req approveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Cuerpo JSON inválido", nil)
		return
	}

	res, err := h.approveUC.Execute(r.Context(), usecase.ApproveClosingRequest{
		ClosingID:  id,
		Version:    req.Version,
		ClientNote: req.ClientNote,
		Batch:      false,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, approveResponse{
		Closing: res.Closing,
		Warning: res.Warning,
	})
}

// approveBatch maneja POST /api/v1/accountant/periods/{period}/closings/approve-batch (spec 010 criterio 3).
func (h *closingsHandler) approveBatch(w http.ResponseWriter, r *http.Request) {
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusUnprocessableEntity, "INVALID_PERIOD", "Formato de período inválido (esperado YYYY-MM)", nil)
		return
	}

	var req usecase.ApproveBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Cuerpo JSON inválido", nil)
		return
	}

	res, err := h.approveBatchUC.Execute(r.Context(), p, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// runReminders maneja POST /api/v1/accountant/reminders/run (spec 010 criterio 11).
func (h *closingsHandler) runReminders(w http.ResponseWriter, r *http.Request) {
	now := h.clock.Now()
	count, err := h.sendRemindersUC.Execute(r.Context(), now)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"reminded": count})
}

