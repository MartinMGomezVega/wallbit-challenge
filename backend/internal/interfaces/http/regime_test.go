package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

type dummyExclusionReader struct {
	exclDate *time.Time
}

func (d *dummyExclusionReader) LatestExclusionDate(_ context.Context, _ string) (*time.Time, error) {
	return d.exclDate, nil
}

func TestRegimeChangeHandler(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, period.Location())
	clk := clock.NewFixed(now)
	clientRepo := memory.NewClientRepository()

	c1 := client.Client{
		ID:        "cli_mono",
		Name:      "Monotributista Test",
		Regime:    client.RegimeMonotributo,
		CreatedAt: time.Date(2025, 10, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = clientRepo.Save(ctx, c1)

	exclDate := time.Date(2026, 10, 14, 0, 0, 0, 0, period.Location())
	exclReader := &dummyExclusionReader{exclDate: &exclDate}

	handler := NewRouter(Deps{
		Clock:                     clk,
		ClientRepo:                clientRepo,
		LatestExclusionDateReader: exclReader,
	})

	t.Run("403 Forbidden si no es contador", func(t *testing.T) {
		body := bytes.NewBufferString(`{"to":"RI","effectiveDate":"2026-10-14"}`)
		req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_mono/regime-change", body)
		req.Header.Set("X-Demo-User", "client:cli_mono")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
	})

	t.Run("404 Not Found si cliente no existe", func(t *testing.T) {
		body := bytes.NewBufferString(`{"to":"RI","effectiveDate":"2026-10-14"}`)
		req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_none/regime-change", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("422 Unprocessable Entity si fecha no coincide y force = false", func(t *testing.T) {
		body := bytes.NewBufferString(`{"to":"RI","effectiveDate":"2026-10-20","force":false}`)
		req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_mono/regime-change", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", w.Code)
		}
		var errResp errorBody
		_ = json.NewDecoder(w.Body).Decode(&errResp)
		if errResp.Error.Code != "EFFECTIVE_DATE_MISMATCH" {
			t.Errorf("code = %s, want EFFECTIVE_DATE_MISMATCH", errResp.Error.Code)
		}
	})

	t.Run("200 OK con fecha coincidente", func(t *testing.T) {
		body := bytes.NewBufferString(`{"to":"RI","effectiveDate":"2026-10-14","note":"Exclusión tope K"}`)
		req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_mono/regime-change", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
		}
		var updated client.Client
		_ = json.NewDecoder(w.Body).Decode(&updated)
		if updated.Regime != client.RegimeRI {
			t.Errorf("updated regime = %v, want RI", updated.Regime)
		}
	})

	t.Run("409 Conflict si ya es RI", func(t *testing.T) {
		body := bytes.NewBufferString(`{"to":"RI","effectiveDate":"2026-10-14"}`)
		req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_mono/regime-change", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("status = %d, want 409", w.Code)
		}
	})

	t.Run("422 si to no es RI o fecha inválida", func(t *testing.T) {
		body := bytes.NewBufferString(`{"to":"OTHER","effectiveDate":"2026-10-14"}`)
		req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_mono/regime-change", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", w.Code)
		}

		body2 := bytes.NewBufferString(`{"to":"RI","effectiveDate":"invalid-date"}`)
		req2 := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_mono/regime-change", body2)
		req2.Header.Set("X-Demo-User", "accountant")
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)

		if w2.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", w2.Code)
		}
	})
}

func TestUpdateRIProfileHandler(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, period.Location())
	clk := clock.NewFixed(now)
	clientRepo := memory.NewClientRepository()

	c1 := client.Client{
		ID:        "cli_ri",
		Regime:    client.RegimeRI,
		CreatedAt: time.Date(2025, 10, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = clientRepo.Save(ctx, c1)

	handler := NewRouter(Deps{
		Clock:      clk,
		ClientRepo: clientRepo,
	})

	t.Run("200 OK actualización válida", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"selfEmployedMonthlyContribution": {"amount": 5000000, "currency": "ARS"},
			"previousVatCredit": {"amount": 1000000, "currency": "ARS"},
			"incomeTaxAdvances": [
				{"period": "2026-10", "amount": {"amount": 15000000, "currency": "ARS"}, "dueDate": "2026-11-15T00:00:00Z"}
			]
		}`)
		req := httptest.NewRequest("PUT", "/api/v1/accountant/clients/cli_ri/ri-profile", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
		}
	})

	t.Run("403 Forbidden si no es contador", func(t *testing.T) {
		body := bytes.NewBufferString(`{"selfEmployedMonthlyContribution": {"amount": 5000000, "currency": "ARS"}}`)
		req := httptest.NewRequest("PUT", "/api/v1/accountant/clients/cli_ri/ri-profile", body)
		req.Header.Set("X-Demo-User", "client:cli_mono")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
	})

	t.Run("404 Not Found si cliente no existe", func(t *testing.T) {
		body := bytes.NewBufferString(`{"selfEmployedMonthlyContribution": {"amount": 5000000, "currency": "ARS"}}`)
		req := httptest.NewRequest("PUT", "/api/v1/accountant/clients/cli_none/ri-profile", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("422 Unprocessable Entity si período inválido o importe negativo", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"selfEmployedMonthlyContribution": {"amount": -100, "currency": "ARS"}
		}`)
		req := httptest.NewRequest("PUT", "/api/v1/accountant/clients/cli_ri/ri-profile", body)
		req.Header.Set("X-Demo-User", "accountant")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", w.Code)
		}

		body2 := bytes.NewBufferString(`{
			"incomeTaxAdvances": [{"period": "invalid-period", "amount": {"amount": 100, "currency": "ARS"}}]
		}`)
		req2 := httptest.NewRequest("PUT", "/api/v1/accountant/clients/cli_ri/ri-profile", body2)
		req2.Header.Set("X-Demo-User", "accountant")
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)

		if w2.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", w2.Code)
		}
	})
}
