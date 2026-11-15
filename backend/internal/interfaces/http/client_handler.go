package http

import (
	"encoding/json"
	"net/http"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
)

type clientHandler struct {
	getMeUC                *usecase.GetMe
	listMyClosingsUC       *usecase.ListMyClosings
	markClosingPaidUC      *usecase.MarkClosingPaid
	listMyNotificationsUC  *usecase.ListMyNotifications
	markNotificationReadUC *usecase.MarkNotificationRead
}

// GET /v1/clients/{clientId}
func (h *clientHandler) getMe(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
	me, err := h.getMeUC.Execute(r.Context(), clientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, me)
}

// GET /v1/clients/{clientId}/closings
func (h *clientHandler) listMyClosings(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
	closings, err := h.listMyClosingsUC.Execute(r.Context(), clientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, closings)
}

// PUT /v1/closings/{closingId}/payment
func (h *clientHandler) markPaid(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	closingID := r.PathValue("closingId")
	callerClientID := id.ClientID
	if id.Role == RoleAccountant {
		callerClientID = ""
	}
	c, err := h.markClosingPaidUC.Execute(r.Context(), closingID, callerClientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"closingId": c.ID,
		"paidAt":    c.PaidAt,
	})
}

// GET /v1/clients/{clientId}/notifications
func (h *clientHandler) listNotifications(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
	notifs, err := h.listMyNotificationsUC.Execute(r.Context(), clientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notifs)
}

// PATCH /v1/clients/{clientId}/notifications/{notificationId}
func (h *clientHandler) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
	notifID := r.PathValue("notificationId")

	var req struct {
		Read bool `json:"read"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !req.Read {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Cuerpo JSON inválido, se requiere {\"read\": true}", nil)
		return
	}

	if err := h.markNotificationReadUC.Execute(r.Context(), notifID, clientID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     notifID,
		"status": "read",
	})
}
