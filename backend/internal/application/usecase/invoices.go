package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ImportResult agrupa el resultado de importar facturas desde CSV (spec 003 criterio 8).
type ImportResult struct {
	Imported          int                 `json:"imported"`
	SkippedDuplicates int                 `json:"skippedDuplicates"`
	Warnings          []invoice.LineIssue `json:"warnings"`
	Errors            []invoice.LineIssue `json:"errors"`
}

// ImportInvoicesCSV procesa un CSV de Mis Comprobantes e importa comprobantes válidos y no duplicados.
type ImportInvoicesCSV struct {
	Clients  ports.ClientRepository
	Invoices ports.InvoiceRepository
	IDGen    ports.IDGenerator
	Clock    ports.Clock
	Notifier ports.ChangeNotifier
	Tx       ports.TxManager
}

// Execute parsea, valida y persiste comprobantes del CSV.
func (uc *ImportInvoicesCSV) Execute(ctx context.Context, clientID string, r io.Reader) (ImportResult, error) {
	rows, warnings, lineErrors, err := invoice.ParseMisComprobantes(r)
	if err != nil {
		return ImportResult{}, domain.NewError(domain.ErrValidation, "INVALID_CSV_HEADERS", err.Error())
	}

	imported := 0
	skippedDuplicates := 0
	affectedPeriods := make(map[string]period.Period)

	err = withinTx(ctx, uc.Tx, func(ctx context.Context) error {
		c, err := uc.Clients.Get(ctx, clientID)
		if err != nil {
			return err
		}

		for _, row := range rows {
			// Validar si el tipo es permitido para el régimen del cliente en la fecha de emisión
			regime := c.RegimeAt(row.IssueDate)
			if regime == client.RegimeMonotributo && row.Type != invoice.TypeC && row.Type != invoice.TypeE {
				lineErrors = append(lineErrors, invoice.LineIssue{
					Line:    row.LineNumber,
					Message: fmt.Sprintf("Comprobante %s no permitido para Monotributo a la fecha de emisión", row.Type),
				})
				continue
			}
			if regime == client.RegimeRI && row.Type != invoice.TypeA && row.Type != invoice.TypeB && row.Type != invoice.TypeE {
				lineErrors = append(lineErrors, invoice.LineIssue{
					Line:    row.LineNumber,
					Message: fmt.Sprintf("Comprobante %s no permitido para Responsable Inscripto a la fecha de emisión", row.Type),
				})
				continue
			}

			key := invoice.Key{
				ClientID:    clientID,
				Type:        row.Type,
				PointOfSale: row.PointOfSale,
				Number:      row.Number,
			}

			exists, err := uc.Invoices.Exists(ctx, clientID, key)
			if err != nil {
				return err
			}
			if exists {
				skippedDuplicates++
				continue
			}

			inv := invoice.Invoice{
				ID:           uc.IDGen.New("inv"),
				ClientID:     clientID,
				Type:         row.Type,
				PointOfSale:  row.PointOfSale,
				Number:       row.Number,
				IssueDate:    row.IssueDate,
				Currency:     row.Currency,
				NetAmount:    row.NetAmount,
				VatAmount:    row.VatAmount,
				TotalAmount:  row.TotalAmount,
				FXRate:       row.FXRate,
				ReceiverName: row.ReceiverName,
				ReceiverDoc:  row.ReceiverDoc,
				Source:       invoice.SourceCSV,
				CreatedAt:    uc.Clock.Now(),
			}

			if err := uc.Invoices.Save(ctx, inv); err != nil {
				return err
			}

			imported++
			p := period.FromTime(inv.IssueDate)
			affectedPeriods[p.String()] = p
		}
		return nil
	})
	if err != nil {
		return ImportResult{}, err
	}

	// Notificar recálculo para cada período afectado
	if uc.Notifier != nil {
		for _, p := range affectedPeriods {
			_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
				ClientID: clientID,
				Period:   p,
			})
		}
	}

	return ImportResult{
		Imported:          imported,
		SkippedDuplicates: skippedDuplicates,
		Warnings:          warnings,
		Errors:            lineErrors,
	}, nil
}

// AddInvoice da de alta una factura emitida manualmente por el cliente.
type AddInvoice struct {
	Clients  ports.ClientRepository
	Invoices ports.InvoiceRepository
	IDGen    ports.IDGenerator
	Clock    ports.Clock
	Notifier ports.ChangeNotifier
}

// Execute valida y guarda una factura manual.
func (uc *AddInvoice) Execute(ctx context.Context, inv invoice.Invoice) (invoice.Invoice, error) {
	c, err := uc.Clients.Get(ctx, inv.ClientID)
	if err != nil {
		return invoice.Invoice{}, err
	}

	regime := c.RegimeAt(inv.IssueDate)
	now := uc.Clock.Now()
	if err := invoice.ValidateManual(inv, string(regime), now); err != nil {
		return invoice.Invoice{}, err
	}

	exists, err := uc.Invoices.Exists(ctx, inv.ClientID, inv.Key())
	if err != nil {
		return invoice.Invoice{}, err
	}
	if exists {
		return invoice.Invoice{}, domain.NewError(domain.ErrConflict, "INVOICE_ALREADY_EXISTS", "Ya existe una factura con este tipo, punto de venta y número")
	}

	inv.ID = uc.IDGen.New("inv")
	inv.Source = invoice.SourceManual
	inv.CreatedAt = now

	if err := uc.Invoices.Save(ctx, inv); err != nil {
		return invoice.Invoice{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: inv.ClientID,
			Period:   period.FromTime(inv.IssueDate),
		})
	}

	return inv, nil
}

// ListInvoices consulta facturas de un cliente para un período dado o histórico completo.
type ListInvoices struct {
	Invoices ports.InvoiceRepository
}

// Execute lista facturas del período indicado.
func (uc *ListInvoices) Execute(ctx context.Context, clientID string, p *period.Period) ([]invoice.Invoice, error) {
	if p != nil {
		return uc.Invoices.ByClientBetween(ctx, clientID, p.Start(), p.End())
	}
	return uc.Invoices.ByClient(ctx, clientID)
}

// AddPurchaseInvoiceInput datos para dar de alta una factura de compra.
type AddPurchaseInvoiceInput struct {
	ClientID     string
	IssueDate    time.Time
	SupplierName string
	SupplierCUIT string
	NetAmount    money.Money
	VatAmount    money.Money
}

// AddPurchaseInvoice da de alta una factura de compra para clientes RI (spec 003 criterio 12).
type AddPurchaseInvoice struct {
	Clients          ports.ClientRepository
	PurchaseInvoices ports.PurchaseInvoiceRepository
	IDGen            ports.IDGenerator
	Clock            ports.Clock
	Notifier         ports.ChangeNotifier
}

// Execute valida régimen RI y campos, persiste y notifica el cambio.
func (uc *AddPurchaseInvoice) Execute(ctx context.Context, input AddPurchaseInvoiceInput) (invoice.PurchaseInvoice, error) {
	c, err := uc.Clients.Get(ctx, input.ClientID)
	if err != nil {
		return invoice.PurchaseInvoice{}, err
	}

	if c.RegimeAt(input.IssueDate) != client.RegimeRI {
		return invoice.PurchaseInvoice{}, domain.NewError(domain.ErrForbidden, "NOT_ALLOWED_FOR_REGIME", "Las facturas de compra solo aplican a Responsables Inscriptos")
	}

	cuit, err := client.ParseCUIT(input.SupplierCUIT)
	if err != nil {
		return invoice.PurchaseInvoice{}, domain.ValidationErrors{
			{Field: "supplierCuit", Msg: "CUIT de proveedor inválido"},
		}
	}

	now := uc.Clock.Now()
	if input.IssueDate.After(now) {
		return invoice.PurchaseInvoice{}, domain.ValidationErrors{
			{Field: "issueDate", Msg: "la fecha de emisión no puede ser futura"},
		}
	}
	if input.SupplierName == "" {
		return invoice.PurchaseInvoice{}, domain.ValidationErrors{
			{Field: "supplierName", Msg: "razón social de proveedor requerida"},
		}
	}

	pi := invoice.PurchaseInvoice{
		ID:           uc.IDGen.New("pi"),
		ClientID:     input.ClientID,
		IssueDate:    input.IssueDate,
		SupplierName: input.SupplierName,
		SupplierCUIT: string(cuit),
		NetAmount:    input.NetAmount,
		VatAmount:    input.VatAmount,
		CreatedAt:    now,
	}

	if err := uc.PurchaseInvoices.Save(ctx, pi); err != nil {
		return invoice.PurchaseInvoice{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: input.ClientID,
			Period:   period.FromTime(input.IssueDate),
		})
	}

	return pi, nil
}

// SetExternalIncome registra los ingresos externos o declara no haber tenido (criterio 13).
type SetExternalIncome struct {
	Inputs   ports.ClientInputRepository
	Clock    ports.Clock
	Notifier ports.ChangeNotifier
}

// SetExternalIncomeInput parámetros para guardar ingresos externos.
type SetExternalIncomeInput struct {
	ClientID string
	Period   period.Period
	Items    []clientinput.ExternalIncomeItem
	None     bool
}

// Execute actualiza los ingresos externos del cliente para el período.
func (uc *SetExternalIncome) Execute(ctx context.Context, input SetExternalIncomeInput) (clientinput.ClientPeriodInput, error) {
	existing, err := uc.Inputs.Get(ctx, input.ClientID, input.Period)
	if err != nil && !errorsIsNotFound(err) {
		return clientinput.ClientPeriodInput{}, err
	}
	if errorsIsNotFound(err) {
		existing = clientinput.ClientPeriodInput{
			ClientID: input.ClientID,
			Period:   input.Period,
		}
	}

	if input.None {
		existing.DeclaredNoExternalIncome = true
		existing.ExternalIncome = nil
	} else {
		existing.DeclaredNoExternalIncome = false
		existing.ExternalIncome = input.Items
	}
	existing.UpdatedAt = uc.Clock.Now()

	if err := uc.Inputs.Save(ctx, existing); err != nil {
		return clientinput.ClientPeriodInput{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: input.ClientID,
			Period:   input.Period,
		})
	}

	return existing, nil
}

// ConfirmPeriodInput sella la confirmación de datos del cliente para el cierre (criterio 14).
type ConfirmPeriodInput struct {
	Inputs   ports.ClientInputRepository
	Clock    ports.Clock
	Notifier ports.ChangeNotifier
}

// Execute confirma el período si ya se respondieron los ingresos externos.
func (uc *ConfirmPeriodInput) Execute(ctx context.Context, clientID string, p period.Period) (clientinput.ClientPeriodInput, error) {
	existing, err := uc.Inputs.Get(ctx, clientID, p)
	if err != nil {
		if errorsIsNotFound(err) {
			return clientinput.ClientPeriodInput{}, clientinput.ErrExternalIncomeNotAnswered
		}
		return clientinput.ClientPeriodInput{}, err
	}

	now := uc.Clock.Now()
	if err := existing.Confirm(now); err != nil {
		return clientinput.ClientPeriodInput{}, err
	}

	if err := uc.Inputs.Save(ctx, existing); err != nil {
		return clientinput.ClientPeriodInput{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: clientID,
			Period:   p,
		})
	}

	return existing, nil
}

func errorsIsNotFound(err error) bool {
	return domain.ErrNotFound == err || (err != nil && err.Error() == domain.ErrNotFound.Error())
}
