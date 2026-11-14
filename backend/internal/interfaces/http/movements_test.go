package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func TestMovementClassificationHTTP(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 10, 0, 0, 0, period.Location())
	clk := clock.NewFixed(now)

	movRepo := memory.NewMovementRepository()
	classRepo := memory.NewClassificationRepository()

	// Seed one movement
	m := movement.Movement{
		ID:         "mov_test_01",
		ClientID:   "cli_demo_01",
		Type:       movement.TypeInboundTransfer,
		Amount:     money.New(150000, money.USD),
		FXRate:     money.Rate(15000000),
		OccurredAt: time.Date(2026, 10, 15, 12, 0, 0, 0, period.Location()),
	}
	if err := movRepo.Save(ctx, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	handler := NewRouter(Deps{
		Clock:              clk,
		MovementRepo:       movRepo,
		ClassificationRepo: classRepo,
	})

	t.Run("200: Successful override by accountant", func(t *testing.T) {
		body := `{"kind":"INCOME","note":"Honorarios confirmados"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/accountant/movements/mov_test_01/classification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "accountant")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200. Body: %s", rec.Code, rec.Body.String())
		}

		var res ClassificationDTO
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if res.MovementID != "mov_test_01" || res.Kind != "INCOME" || res.Rule != "MANUAL_OVERRIDE" || res.OverriddenBy != "accountant" {
			t.Fatalf("unexpected response: %+v", res)
		}
		if res.Note != "Honorarios confirmados" {
			t.Fatalf("unexpected note: %s", res.Note)
		}
	})

	t.Run("404: Movement not found", func(t *testing.T) {
		body := `{"kind":"INCOME"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/accountant/movements/mov_non_existent/classification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "accountant")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("422: UNCLASSIFIED is rejected", func(t *testing.T) {
		body := `{"kind":"UNCLASSIFIED"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/accountant/movements/mov_test_01/classification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "accountant")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("403: Forbidden for client role", func(t *testing.T) {
		body := `{"kind":"INCOME"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/accountant/movements/mov_test_01/classification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "client:cli_demo_01")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("401: Unauthorized without identity header", func(t *testing.T) {
		body := `{"kind":"INCOME"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/accountant/movements/mov_test_01/classification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("400: Malformed JSON", func(t *testing.T) {
		body := `{bad json`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/accountant/movements/mov_test_01/classification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "accountant")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}
