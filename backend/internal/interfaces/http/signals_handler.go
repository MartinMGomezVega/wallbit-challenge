package http

import (
	"net/http"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
)

func signalsCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, signal.Catalog())
}

