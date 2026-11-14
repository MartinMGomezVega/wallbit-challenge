package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
)

func TestSignalsCatalogEndpoint(t *testing.T) {
	clk := clock.NewFixed(time.Now())
	handler := NewRouter(Deps{Clock: clk})

	t.Run("403 Forbidden si no es contador", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/accountant/signals/catalog", nil)
		req.Header.Set("X-Demo-User", "client:cli_1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
	})

	t.Run("200 OK devuelve los 13 ítems del catálogo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/accountant/signals/catalog", nil)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}

		var items []signal.CatalogItem
		if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		if len(items) != 13 {
			t.Errorf("len(items) = %d, want 13", len(items))
		}
		if items[0].Code != signal.CodeMonoExclusionKCapExceeded {
			t.Errorf("items[0].Code = %s, want MONO_EXCLUSION_K_CAP_EXCEEDED", items[0].Code)
		}
	})
}

