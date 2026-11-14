package http

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func TestParametersEndpoints(t *testing.T) {
	repo := memory.NewParametersRepository()
	clk := clock.NewFixed(testNow)
	deps := Deps{
		Clock:          clk,
		CORSOrigin:     "http://localhost:5173",
		ParametersRepo: repo,
	}
	router := NewRouter(deps)

	// 1. GET /api/v1/accountant/parameters sin auth -> 401
	rec := do(t, router, "GET", "/api/v1/accountant/parameters", "", nil)
	if rec.Code != 401 {
		t.Fatalf("esperaba 401, got %d", rec.Code)
	}

	// 2. GET con rol cliente -> 403
	rec = do(t, router, "GET", "/api/v1/accountant/parameters", "client:cli_1", nil)
	if rec.Code != 403 {
		t.Fatalf("esperaba 403, got %d", rec.Code)
	}

	// 3. GET con rol accountant -> 200 defaults
	rec = do(t, router, "GET", "/api/v1/accountant/parameters", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("esperaba 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var getDTO ParametersDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &getDTO); err != nil {
		t.Fatal(err)
	}
	if !getDTO.InvoiceECountsForCap || getDTO.FXRule != "BNA_BUY_PREV_BUSINESS_DAY" || getDTO.CategoryCapWarningBp != 8000 {
		t.Fatalf("defaults inesperados: %+v", getDTO)
	}

	// 4. PUT válido -> 200
	getDTO.CategoryCapWarningBp = 8500
	getDTO.FXRule = "BNA_SELL_PREV_BUSINESS_DAY"
	payload, _ := json.Marshal(getDTO)

	rec = do(t, router, "PUT", "/api/v1/accountant/parameters", "accountant", bytes.NewReader(payload))
	if rec.Code != 200 {
		t.Fatalf("esperaba 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var putResp putParametersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &putResp); err != nil {
		t.Fatal(err)
	}
	if putResp.Parameters.CategoryCapWarningBp != 8500 || putResp.Parameters.FXRule != "BNA_SELL_PREV_BUSINESS_DAY" {
		t.Fatalf("valores no actualizados: %+v", putResp)
	}
	if putResp.Parameters.UpdatedBy != "accountant" || putResp.Parameters.UpdatedAt == "" {
		t.Fatalf("auditoría incompleta: %+v", putResp)
	}
	if putResp.Recalculated < 0 {
		t.Fatalf("recalculated inválido: %d", putResp.Recalculated)
	}

	// 5. PUT con validación errónea (bp > 100000) -> 422
	getDTO.CategoryCapWarningBp = 120000
	badPayload, _ := json.Marshal(getDTO)
	rec = do(t, router, "PUT", "/api/v1/accountant/parameters", "accountant", bytes.NewReader(badPayload))
	if rec.Code != 422 {
		t.Fatalf("esperaba 422, got %d (%s)", rec.Code, rec.Body.String())
	}
	errPayload := decodeErr(t, rec)
	if errPayload.Code != "VALIDATION_ERROR" || errPayload.Fields["categoryCapWarningBp"] == "" {
		t.Fatalf("error payload inesperado: %+v", errPayload)
	}

	// 6. PUT con JSON malformado -> 400
	rec = do(t, router, "PUT", "/api/v1/accountant/parameters", "accountant", bytes.NewReader([]byte("{invalid-json")))
	if rec.Code != 400 {
		t.Fatalf("esperaba 400, got %d", rec.Code)
	}
}

func TestParametersRepoMemory(t *testing.T) {
	repo := memory.NewParametersRepository()
	ctx := context.Background()
	p, err := repo.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	p.KCapWarningBp = 9000
	if err := repo.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	updated, err := repo.Get(ctx)
	if err != nil || updated.KCapWarningBp != 9000 {
		t.Fatalf("repo memory save/get falló: %+v", updated)
	}
}
