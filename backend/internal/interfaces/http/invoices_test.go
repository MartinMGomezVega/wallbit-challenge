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

	t.Run("POST /v1/clients/{clientId}/invoice-imports: Success, Idempotency & 422", func(t *testing.T) {
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

			req := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/invoice-imports", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("X-Demo-User", "client:cli_mono")
			return req
		}

		// 1. Primera importación: imported = 1
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, makeMultipart(csvContent))
		if rec1.Code != http.StatusCreated {
			t.Fatalf("first import status = %d, want 201. Body: %s", rec1.Code, rec1.Body.String())
		}
		var res1 map[string]any
		_ = json.NewDecoder(rec1.Body).Decode(&res1)
		if res1["imported"] != float64(1) || res1["skippedDuplicates"] != float64(0) {
			t.Fatalf("unexpected res1: %+v", res1)
		}

		// 2. Segunda importación (idempotente): imported = 0, skippedDuplicates = 1
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, makeMultipart(csvContent))
		if rec2.Code != http.StatusCreated {
			t.Fatalf("second import status = %d, want 201", rec2.Code)
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

		// 4. Role other client trying to call -> 403
		reqOther := makeMultipart(csvContent)
		reqOther.Header.Set("X-Demo-User", "client:cli_ri")
		recOther := httptest.NewRecorder()
		handler.ServeHTTP(recOther, reqOther)
		if recOther.Code != http.StatusForbidden {
			t.Fatalf("other client status = %d, want 403", recOther.Code)
		}
	})

	t.Run("POST /v1/clients/{clientId}/invoices: Manual, Validation, Conflict", func(t *testing.T) {
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
		req := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/invoices", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", "client:cli_mono")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("POST /v1/clients/{clientId}/invoices status = %d, want 201. Body: %s", rec.Code, rec.Body.String())
		}

		// Duplicado -> 409
		reqDup := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/invoices", bytes.NewBufferString(body))
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
		reqInv := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/invoices", bytes.NewBufferString(bodyInvalid))
		reqInv.Header.Set("Content-Type", "application/json")
		reqInv.Header.Set("X-Demo-User", "client:cli_mono")
		recInv := httptest.NewRecorder()
		handler.ServeHTTP(recInv, reqInv)
		if recInv.Code != http.StatusUnprocessableEntity {
			t.Fatalf("invalid invoice status = %d, want 422", recInv.Code)
		}
	})

	t.Run("GET /v1/clients/{clientId}/invoices (client & accountant)", func(t *testing.T) {
		// Client list
		reqClient := httptest.NewRequest(http.MethodGet, "/v1/clients/cli_mono/invoices?period=2026-10", nil)
		reqClient.Header.Set("X-Demo-User", "client:cli_mono")
		recClient := httptest.NewRecorder()
		handler.ServeHTTP(recClient, reqClient)
		if recClient.Code != http.StatusOK {
			t.Fatalf("GET /v1/clients/{clientId}/invoices status = %d, want 200", recClient.Code)
		}
		var listClient []InvoiceDTO
		_ = json.NewDecoder(recClient.Body).Decode(&listClient)
		if len(listClient) == 0 {
			t.Fatal("expected invoices in /v1/clients/{clientId}/invoices")
		}

		// Accountant list
		reqAcc := httptest.NewRequest(http.MethodGet, "/v1/clients/cli_mono/invoices?period=2026-10", nil)
		reqAcc.Header.Set("X-Demo-User", "accountant")
		recAcc := httptest.NewRecorder()
		handler.ServeHTTP(recAcc, reqAcc)
		if recAcc.Code != http.StatusOK {
			t.Fatalf("GET /v1/clients/{clientId}/invoices (accountant) status = %d, want 200", recAcc.Code)
		}
		var listAcc []InvoiceDTO
		_ = json.NewDecoder(recAcc.Body).Decode(&listAcc)
		if len(listAcc) != len(listClient) {
			t.Fatalf("accountant got %d, client got %d", len(listAcc), len(listClient))
		}
	})

	t.Run("POST /v1/clients/{clientId}/purchase-invoices: RI only", func(t *testing.T) {
		body := `{
			"issueDate": "2026-10-18",
			"supplierName": "Proveedor Servicios SRL",
			"supplierCuit": "20-35000000-4",
			"netAmount": { "amount": "50000.00", "currency": "ARS" },
			"vatAmount": { "amount": "10500.00", "currency": "ARS" }
		}`

		// 1. Monotributo -> 403 NOT_ALLOWED_FOR_REGIME
		reqMono := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/purchase-invoices", bytes.NewBufferString(body))
		reqMono.Header.Set("Content-Type", "application/json")
		reqMono.Header.Set("X-Demo-User", "client:cli_mono")
		recMono := httptest.NewRecorder()
		handler.ServeHTTP(recMono, reqMono)
		if recMono.Code != http.StatusForbidden {
			t.Fatalf("purchase invoice Monotributo status = %d, want 403", recMono.Code)
		}

		// 2. RI -> 201 Created
		reqRI := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_ri/purchase-invoices", bytes.NewBufferString(body))
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
		reqConfirmEarly := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-10/confirmation", nil)
		reqConfirmEarly.Header.Set("X-Demo-User", "client:cli_mono")
		recConfirmEarly := httptest.NewRecorder()
		handler.ServeHTTP(recConfirmEarly, reqConfirmEarly)
		if recConfirmEarly.Code != http.StatusUnprocessableEntity {
			t.Fatalf("confirm before answering status = %d, want 422", recConfirmEarly.Code)
		}

		// 2. Set external income (none: true) -> 200
		bodyNone := `{"none": true}`
		reqIncome := httptest.NewRequest(http.MethodPut, "/v1/clients/cli_mono/periods/2026-10/external-income", bytes.NewBufferString(bodyNone))
		reqIncome.Header.Set("Content-Type", "application/json")
		reqIncome.Header.Set("X-Demo-User", "client:cli_mono")
		recIncome := httptest.NewRecorder()
		handler.ServeHTTP(recIncome, reqIncome)
		if recIncome.Code != http.StatusOK {
			t.Fatalf("set external income status = %d, want 200", recIncome.Code)
		}

		// 3. Confirm now -> 201
		reqConfirm := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-10/confirmation", nil)
		reqConfirm.Header.Set("X-Demo-User", "client:cli_mono")
		recConfirm := httptest.NewRecorder()
		handler.ServeHTTP(recConfirm, reqConfirm)
		if recConfirm.Code != http.StatusCreated {
			t.Fatalf("confirm status = %d, want 201", recConfirm.Code)
		}
	})

	t.Run("POST /v1/clients/{clientId}/invoice-imports accountant", func(t *testing.T) {
		csvContent := `Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total
15/10/2026,19 - Factura E,00001,00000077,00000077,74392819289901,80,DE329182391,Brightloop GmbH,1482.15,USD,500.00,0.00,0.00,0.00,500.00
`
		makeMultipart := func(content string, user string) *http.Request {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "mis_comprobantes.csv")
			_, _ = part.Write([]byte(content))
			_ = writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/invoice-imports", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("X-Demo-User", user)
			return req
		}

		// 1. Rol cliente ajeno intentando acceder -> 403 Forbidden
		reqForbidden := makeMultipart(csvContent, "client:cli_ri")
		recForbidden := httptest.NewRecorder()
		handler.ServeHTTP(recForbidden, reqForbidden)
		if recForbidden.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for client role, got %d", recForbidden.Code)
		}

		// 2. Rol contador con cliente inexistente -> 404
		body404 := &bytes.Buffer{}
		w404 := multipart.NewWriter(body404)
		p404, _ := w404.CreateFormFile("file", "test.csv")
		_, _ = p404.Write([]byte(csvContent))
		_ = w404.Close()
		req404 := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_inexistente/invoice-imports", body404)
		req404.Header.Set("Content-Type", w404.FormDataContentType())
		req404.Header.Set("X-Demo-User", "accountant")
		rec404 := httptest.NewRecorder()
		handler.ServeHTTP(rec404, req404)
		if rec404.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for non-existent client, got %d: %s", rec404.Code, rec404.Body.String())
		}

		// 3. Rol contador éxito -> 201
		reqOK := makeMultipart(csvContent, "accountant")
		recOK := httptest.NewRecorder()
		handler.ServeHTTP(recOK, reqOK)
		if recOK.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", recOK.Code, recOK.Body.String())
		}
		var res map[string]any
		_ = json.Unmarshal(recOK.Body.Bytes(), &res)
		if res["imported"] != float64(1) {
			t.Errorf("imported = %v, want 1", res["imported"])
		}

		// 4. Archivo inválido -> 422
		reqBad := makeMultipart("cabecera,invalida\n1,2", "accountant")
		recBad := httptest.NewRecorder()
		handler.ServeHTTP(recBad, reqBad)
		if recBad.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for invalid CSV headers, got %d", recBad.Code)
		}
	})

	t.Run("POST /v1/clients/{clientId}/periods/{period}/confirmation accountant", func(t *testing.T) {
		// 1. Rol cliente ajeno -> 403 Forbidden
		bodyValid := `{"externalIncome": {"none": true}, "note": "Confirmado por llamada"}`
		reqForbidden := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-11/confirmation", bytes.NewBufferString(bodyValid))
		reqForbidden.Header.Set("Content-Type", "application/json")
		reqForbidden.Header.Set("X-Demo-User", "client:cli_ri")
		recForbidden := httptest.NewRecorder()
		handler.ServeHTTP(recForbidden, reqForbidden)
		if recForbidden.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for client role, got %d", recForbidden.Code)
		}

		// 2. Nota vacía -> 422 NOTE_REQUIRED
		bodyNoNote := `{"externalIncome": {"none": true}, "note": "   "}`
		reqNoNote := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-11/confirmation", bytes.NewBufferString(bodyNoNote))
		reqNoNote.Header.Set("Content-Type", "application/json")
		reqNoNote.Header.Set("X-Demo-User", "accountant")
		recNoNote := httptest.NewRecorder()
		handler.ServeHTTP(recNoNote, reqNoNote)
		if recNoNote.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for missing note, got %d", recNoNote.Code)
		}

		// 3. Confirmación exitosa por contador -> 201
		reqOK := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-11/confirmation", bytes.NewBufferString(bodyValid))
		reqOK.Header.Set("Content-Type", "application/json")
		reqOK.Header.Set("X-Demo-User", "accountant")
		recOK := httptest.NewRecorder()
		handler.ServeHTTP(recOK, reqOK)
		if recOK.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", recOK.Code, recOK.Body.String())
		}
		var inputRes map[string]any
		_ = json.Unmarshal(recOK.Body.Bytes(), &inputRes)
		if inputRes["confirmedBy"] != "ACCOUNTANT" {
			t.Errorf("confirmedBy = %v, want ACCOUNTANT", inputRes["confirmedBy"])
		}
		if inputRes["confirmationNote"] != "Confirmado por llamada" {
			t.Errorf("confirmationNote = %v, want 'Confirmado por llamada'", inputRes["confirmationNote"])
		}

		// 4. Reintento sobre período ya confirmado -> 409 ALREADY_CONFIRMED
		recConflict := httptest.NewRecorder()
		reqConflict := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-11/confirmation", bytes.NewBufferString(bodyValid))
		reqConflict.Header.Set("Content-Type", "application/json")
		reqConflict.Header.Set("X-Demo-User", "accountant")
		handler.ServeHTTP(recConflict, reqConflict)
		if recConflict.Code != http.StatusConflict {
			t.Fatalf("expected 409 ALREADY_CONFIRMED, got %d", recConflict.Code)
		}

		// 5. Período inválido -> 400
		reqBadP := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/bad-period/confirmation", bytes.NewBufferString(bodyValid))
		reqBadP.Header.Set("Content-Type", "application/json")
		reqBadP.Header.Set("X-Demo-User", "accountant")
		recBadP := httptest.NewRecorder()
		handler.ServeHTTP(recBadP, reqBadP)
		if recBadP.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for bad period, got %d", recBadP.Code)
		}

		// 6. JSON inválido -> 400
		reqBadJSON := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-12/confirmation", bytes.NewBufferString("{not-json"))
		reqBadJSON.Header.Set("Content-Type", "application/json")
		reqBadJSON.Header.Set("X-Demo-User", "accountant")
		recBadJSON := httptest.NewRecorder()
		handler.ServeHTTP(recBadJSON, reqBadJSON)
		if recBadJSON.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for bad json, got %d", recBadJSON.Code)
		}

		// 7. Confirmación con items de ingreso externo (con montos float y string)
		bodyItems := `{"externalIncome": {"items": [{"amount": {"amount": 15000, "currency": "USD"}, "description": "Consultoría"}, {"amount": {"amount": "250.50", "currency": "USD"}, "description": "Soporte"}]}, "note": "Nota válida con items"}`
		reqItems := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/periods/2026-12/confirmation", bytes.NewBufferString(bodyItems))
		reqItems.Header.Set("Content-Type", "application/json")
		reqItems.Header.Set("X-Demo-User", "accountant")
		recItems := httptest.NewRecorder()
		handler.ServeHTTP(recItems, reqItems)
		if recItems.Code != http.StatusCreated {
			t.Fatalf("expected 201 for confirm with items, got %d: %s", recItems.Code, recItems.Body.String())
		}
	})

	t.Run("POST /v1/clients/{clientId}/invoice-imports edge cases", func(t *testing.T) {
		// Falta campo file -> 400
		var buf bytes.Buffer
		mp := multipart.NewWriter(&buf)
		_ = mp.WriteField("other", "value")
		_ = mp.Close()
		reqNoFile := httptest.NewRequest(http.MethodPost, "/v1/clients/cli_mono/invoice-imports", &buf)
		reqNoFile.Header.Set("Content-Type", mp.FormDataContentType())
		reqNoFile.Header.Set("X-Demo-User", "accountant")
		recNoFile := httptest.NewRecorder()
		handler.ServeHTTP(recNoFile, reqNoFile)
		if recNoFile.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing file field, got %d", recNoFile.Code)
		}
	})
}
