package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func TestInvoicesHTTP(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, period.Location())
	clk := clock.NewFixed(now)
	idGen := idgen.NewSequential()

	clientRepo := memory.NewClientRepository()
	invRepo := memory.NewInvoiceRepository()
	piRepo := memory.NewPurchaseInvoiceRepository()
	ciRepo := memory.NewClientInputRepository()

	cMono := client.Client{
		ID:     "cli_mono",
		Name:   "Paula Rojas",
		Regime: client.RegimeMonotributo,
	}
	cRI := client.Client{
		ID:     "cli_ri",
		Name:   "Camila Torres",
		Regime: client.RegimeRI,
	}
	_ = clientRepo.Save(ctx, cMono)
	_ = clientRepo.Save(ctx, cRI)

	handler := NewRouter(Deps{
		Clock:               clk,
		ClientRepo:          clientRepo,
		InvoiceRepo:         invRepo,
		PurchaseInvoiceRepo: piRepo,
		ClientInputRepo:     ciRepo,
		IDGen:               idGen,
	})

	t.Run("POST /me/invoices/import: Success, Idempotency & 422", func(t *testing.T) {
		csvContent := `Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total
12/10/2026,19 - Factura E,00001,00000001,00000001,74392819289901,80,DE329182391,Brightloop GmbH,1482.15,USD,850.00,0.00,0.00,0.00,850.00
`
		// Helper para multipart request
		makeMultipart := func(content string) *http.Request {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "mis_comprobantes.csv")
			_, _ = part.Write([]byte(content))
			_ = writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/me/invoices/import", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("X-Demo-User", "client:cli_mono")
			return req
		}

		// 1. Primera importación: imported = 1
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, makeMultipart(csvContent))
		if rec1.Code != http.StatusOK {
			t.Fatalf("first import status = %d, want 200. Body: %s", rec1.Code, rec1.Body.String())
		}
		var res1 map[string]any
		_ = json.NewDecoder(rec1.Body).Decode(&res1)
		if res1["imported"] != float64(1) || res1["skippedDuplicates"] != float64(0) {
			t.Fatalf("unexpected res1: %+v", res1)
		}

		// 2. Segunda importación (idempotente): imported = 0, skippedDuplicates = 1
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, makeMultipart(csvContent))
		if rec2.Code != http.StatusOK {
			t.Fatalf("second import status = %d, want 200", rec2.Code)
		}
		var res2 map[string]any
		_ = json.NewDecoder(rec2.Body).Decode(&res2)
		if res2["imported"] != float64(0) || res2["skippedDuplicates"] != float64(1) {
			t.Fatalf("unexpected res2: %+v", res2)
		}

		// 3. Encabezados inválidos -> 422
		recBad := httptest.NewRecorder()
		handler.ServeHTTP(recBad, makeMultipart("invalido\n1,2,3"))
		if recBad.Code != http.StatusUnprocessableEntity {
			t.Fatalf("bad headers status = %d, want 422", recBad.Code)
		}

		// 4. Role accountant trying to call /me/invoices/import -> 403
		reqAcc := makeMultipart(csvContent)
		reqAcc.Header.Set("X-Demo-User", "accountant")
		recAcc := httptest.NewRecorder()
		handler.ServeHTTP(recAcc, reqAcc)
		if recAcc.Code != http.StatusForbidden {
			t.Fatalf("accountant role status = %d, want 403", recAcc.Code)
		}
	})

	t.Run("POST /me/invoices: Manual, Validation, Conflict", func(t *testing.T) {
		body := `{
			"type": "E",
			"pointOfSale": 1,
			"number": 50,
			"issueDate": "2026-10-14",
			"currency": "USD",
			"totalAmount": { "amount": "1200.00", "currency": "USD" },
			"netAmount": { "amount": "1200.00", "currency": "USD" },
			"vatAmount": { "amount": "0.00", "currency": "USD" },
			"receiverName": "Kiwi Studio Ltd"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/me/invoices", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "client:cli_mono")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("POST /me/invoices status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}

		// Duplicado -> 409
		reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/me/invoices", bytes.NewBufferString(body))
		reqDup.Header.Set("Content-Type", "application/json")
		reqDup.Header.Set("X-Demo-User", "client:cli_mono")
		recDup := httptest.NewRecorder()
		handler.ServeHTTP(recDup, reqDup)
		if recDup.Code != http.StatusConflict {
			t.Fatalf("duplicate invoice status = %d, want 409", recDup.Code)
		}

		// Error de validación (tipo no permitido para Monotributo: Factura A) -> 422
		bodyInvalid := `{
			"type": "A",
			"pointOfSale": 1,
			"number": 51,
			"issueDate": "2026-10-14",
			"currency": "ARS",
			"totalAmount": { "amount": "1000.00", "currency": "ARS" },
			"netAmount": { "amount": "1000.00", "currency": "ARS" },
			"vatAmount": { "amount": "0.00", "currency": "ARS" },
			"receiverName": "Local SRL"
		}`
		reqInv := httptest.NewRequest(http.MethodPost, "/api/v1/me/invoices", bytes.NewBufferString(bodyInvalid))
		reqInv.Header.Set("Content-Type", "application/json")
		reqInv.Header.Set("X-Demo-User", "client:cli_mono")
		recInv := httptest.NewRecorder()
		handler.ServeHTTP(recInv, reqInv)
		if recInv.Code != http.StatusUnprocessableEntity {
			t.Fatalf("invalid invoice status = %d, want 422", recInv.Code)
		}
	})

	t.Run("GET /me/invoices & GET /accountant/clients/{id}/invoices", func(t *testing.T) {
		// Client list
		reqClient := httptest.NewRequest(http.MethodGet, "/api/v1/me/invoices?period=2026-10", nil)
		reqClient.Header.Set("X-Demo-User", "client:cli_mono")
		recClient := httptest.NewRecorder()
		handler.ServeHTTP(recClient, reqClient)
		if recClient.Code != http.StatusOK {
			t.Fatalf("GET /me/invoices status = %d, want 200", recClient.Code)
		}
		var listClient []InvoiceDTO
		_ = json.NewDecoder(recClient.Body).Decode(&listClient)
		if len(listClient) == 0 {
			t.Fatal("expected invoices in /me/invoices")
		}

		// Accountant list
		reqAcc := httptest.NewRequest(http.MethodGet, "/api/v1/accountant/clients/cli_mono/invoices?period=2026-10", nil)
		reqAcc.Header.Set("X-Demo-User", "accountant")
		recAcc := httptest.NewRecorder()
		handler.ServeHTTP(recAcc, reqAcc)
		if recAcc.Code != http.StatusOK {
			t.Fatalf("GET /accountant/clients/{id}/invoices status = %d, want 200", recAcc.Code)
		}
		var listAcc []InvoiceDTO
		_ = json.NewDecoder(recAcc.Body).Decode(&listAcc)
		if len(listAcc) != len(listClient) {
			t.Fatalf("accountant got %d, client got %d", len(listAcc), len(listClient))
		}
	})

	t.Run("POST /me/purchase-invoices: RI only", func(t *testing.T) {
		body := `{
			"issueDate": "2026-10-18",
			"supplierName": "Proveedor Servicios SRL",
			"supplierCuit": "20-35000000-4",
			"netAmount": { "amount": "50000.00", "currency": "ARS" },
			"vatAmount": { "amount": "10500.00", "currency": "ARS" }
		}`

		// 1. Monotributo -> 403 NOT_ALLOWED_FOR_REGIME
		reqMono := httptest.NewRequest(http.MethodPost, "/api/v1/me/purchase-invoices", bytes.NewBufferString(body))
		reqMono.Header.Set("Content-Type", "application/json")
		reqMono.Header.Set("X-Demo-User", "client:cli_mono")
		recMono := httptest.NewRecorder()
		handler.ServeHTTP(recMono, reqMono)
		if recMono.Code != http.StatusForbidden {
			t.Fatalf("purchase invoice Monotributo status = %d, want 403", recMono.Code)
		}

		// 2. RI -> 201 Created
		reqRI := httptest.NewRequest(http.MethodPost, "/api/v1/me/purchase-invoices", bytes.NewBufferString(body))
		reqRI.Header.Set("Content-Type", "application/json")
		reqRI.Header.Set("X-Demo-User", "client:cli_ri")
		recRI := httptest.NewRecorder()
		handler.ServeHTTP(recRI, reqRI)
		if recRI.Code != http.StatusCreated {
			t.Fatalf("purchase invoice RI status = %d, want 201. Body: %s", recRI.Code, recRI.Body.String())
		}
	})

	t.Run("External income & Confirm", func(t *testing.T) {
		// 1. Confirm before answer -> 422
		reqConfirmEarly := httptest.NewRequest(http.MethodPost, "/api/v1/me/periods/2026-10/confirm", nil)
		reqConfirmEarly.Header.Set("X-Demo-User", "client:cli_mono")
		recConfirmEarly := httptest.NewRecorder()
		handler.ServeHTTP(recConfirmEarly, reqConfirmEarly)
		if recConfirmEarly.Code != http.StatusUnprocessableEntity {
			t.Fatalf("confirm before answering status = %d, want 422", recConfirmEarly.Code)
		}

		// 2. Set external income (none: true) -> 200
		bodyNone := `{"none": true}`
		reqIncome := httptest.NewRequest(http.MethodPut, "/api/v1/me/periods/2026-10/external-income", bytes.NewBufferString(bodyNone))
		reqIncome.Header.Set("Content-Type", "application/json")
		reqIncome.Header.Set("X-Demo-User", "client:cli_mono")
		recIncome := httptest.NewRecorder()
		handler.ServeHTTP(recIncome, reqIncome)
		if recIncome.Code != http.StatusOK {
			t.Fatalf("set external income status = %d, want 200", recIncome.Code)
		}

		// 3. Confirm now -> 200
		reqConfirm := httptest.NewRequest(http.MethodPost, "/api/v1/me/periods/2026-10/confirm", nil)
		reqConfirm.Header.Set("X-Demo-User", "client:cli_mono")
		recConfirm := httptest.NewRecorder()
		handler.ServeHTTP(recConfirm, reqConfirm)
		if recConfirm.Code != http.StatusOK {
			t.Fatalf("confirm status = %d, want 200", recConfirm.Code)
		}
	})
}
