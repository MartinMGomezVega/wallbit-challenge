package http

import (
	"net/http"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
)

type clientHandler struct {
	getMeUC                *usecase.GetMe
	listDemoClientsUC      *usecase.ListDemoClients
	listMyClosingsUC       *usecase.ListMyClosings
	markClosingPaidUC      *usecase.MarkClosingPaid
	listMyNotificationsUC  *usecase.ListMyNotifications
	markNotificationReadUC *usecase.MarkNotificationRead
}

// GET /api/v1/demo/clients
func (h *clientHandler) listDemoClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.listDemoClientsUC.Execute(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clients)
}

// GET /api/v1/me
func (h *clientHandler) getMe(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	me, err := h.getMeUC.Execute(r.Context(), id.ClientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, me)
}

// GET /api/v1/me/closings
func (h *clientHandler) listMyClosings(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	closings, err := h.listMyClosingsUC.Execute(r.Context(), id.ClientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, closings)
}

// POST /api/v1/me/closings/{id}/paid
func (h *clientHandler) markPaid(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	closingID := r.PathValue("id")
	c, err := h.markClosingPaidUC.Execute(r.Context(), closingID, id.ClientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"closingId": c.ID,
		"paidAt":    c.PaidAt,
	})
}

// GET /api/v1/me/notifications
func (h *clientHandler) listNotifications(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	notifs, err := h.listMyNotificationsUC.Execute(r.Context(), id.ClientID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notifs)
}

// POST /api/v1/me/notifications/{id}/read
func (h *clientHandler) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	notifID := r.PathValue("id")
	if err := h.markNotificationReadUC.Execute(r.Context(), notifID, id.ClientID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     notifID,
		"status": "read",
	})
}
