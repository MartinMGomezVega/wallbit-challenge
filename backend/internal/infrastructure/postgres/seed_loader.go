package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LoadSeed carga el dataset semilla y clasificaciones iniciales en Postgres usando CopyFrom.
// Si reset es true, trunca las tablas antes de cargar. Si ya existen clientes y reset es false, es no-op (idempotente).
func LoadSeed(ctx context.Context, pool *pgxpool.Pool, ds *seed.Dataset, classifications []movement.Classification, reset bool) (bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if reset {
		if _, err := tx.Exec(ctx, `
			TRUNCATE TABLE
				approvals, closing_signals, closings,
				external_income_items, client_period_inputs,
				purchase_invoices, invoices,
				classifications, movements,
				income_tax_advances, regime_history, clients,
				notifications, outbox, parameters
			CASCADE
		`); err != nil {
			return false, fmt.Errorf("truncate seed tables: %w", err)
		}
	} else {
		var count int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM clients`).Scan(&count); err != nil {
			return false, fmt.Errorf("count clients: %w", err)
		}
		if count > 0 {
			return false, tx.Commit(ctx)
		}
	}

	if err := copyClients(ctx, tx, ds); err != nil {
		return false, err
	}
	if err := copyMovements(ctx, tx, ds, classifications); err != nil {
		return false, err
	}
	if err := copyInvoices(ctx, tx, ds); err != nil {
		return false, err
	}
	if err := copyClientInputs(ctx, tx, ds); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func copyClients(ctx context.Context, tx pgx.Tx, ds *seed.Dataset) error {
	if _, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"clients"},
		[]string{"id", "name", "email", "cuit", "regime", "category", "ri_self_employed_cents", "ri_prev_vat_credit_cents", "name_search", "created_at"},
		pgx.CopyFromSlice(len(ds.Clients), func(i int) ([]any, error) {
			c := ds.Clients[i]
			var seCents, prevVatCents int64
			if c.RIProfile != nil {
				seCents = c.RIProfile.SelfEmployedMonthlyContribution.Amount
				prevVatCents = c.RIProfile.PreviousVatCredit.Amount
			}
			return []any{
				c.ID, c.Name, c.Email, string(c.CUIT), string(c.Regime), string(c.MonotributoCategory),
				seCents, prevVatCents, strings.ToLower(c.Name), c.CreatedAt,
			}, nil
		}),
	); err != nil {
		return fmt.Errorf("copy clients: %w", err)
	}

	type rhRow struct {
		clientID string
		seq      int
		regime   string
		from     any
		to       any
		reason   string
		note     string
	}
	var rhRows []rhRow
	type advRow struct {
		clientID string
		period   string
		amount   int64
		due      any
	}
	var advRows []advRow

	for _, c := range ds.Clients {
		for seq, h := range c.RegimeHistory {
			rhRows = append(rhRows, rhRow{c.ID, seq, string(h.Regime), h.From, h.To, h.Reason, h.Note})
		}
		if c.RIProfile != nil {
			for _, adv := range c.RIProfile.IncomeTaxAdvances {
				advRows = append(advRows, advRow{c.ID, adv.Period.String(), adv.Amount.Amount, adv.DueDate})
			}
		}
	}

	if len(rhRows) > 0 {
		if _, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"regime_history"},
			[]string{"client_id", "seq", "regime", "valid_from", "valid_to", "reason", "note"},
			pgx.CopyFromSlice(len(rhRows), func(i int) ([]any, error) {
				r := rhRows[i]
				return []any{r.clientID, r.seq, r.regime, r.from, r.to, r.reason, r.note}, nil
			}),
		); err != nil {
			return fmt.Errorf("copy regime_history: %w", err)
		}
	}

	if len(advRows) > 0 {
		if _, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"income_tax_advances"},
			[]string{"client_id", "period", "amount_cents", "due_date"},
			pgx.CopyFromSlice(len(advRows), func(i int) ([]any, error) {
				a := advRows[i]
				return []any{a.clientID, a.period, a.amount, a.due}, nil
			}),
		); err != nil {
			return fmt.Errorf("copy income_tax_advances: %w", err)
		}
	}
	return nil
}

func copyMovements(ctx context.Context, tx pgx.Tx, ds *seed.Dataset, classifications []movement.Classification) error {
	if _, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"movements"},
		[]string{"id", "client_id", "type", "amount_cents", "currency", "fx_rate_scaled", "occurred_at", "cp_name", "cp_country", "cp_kind", "cp_own", "description"},
		pgx.CopyFromSlice(len(ds.Movements), func(i int) ([]any, error) {
			m := ds.Movements[i]
			return []any{
				m.ID, m.ClientID, string(m.Type), m.Amount.Amount, string(m.Amount.Currency),
				int64(m.FXRate), m.OccurredAt, m.Counterparty.Name, m.Counterparty.Country,
				string(m.Counterparty.Kind), m.Counterparty.IsOwnAccount, m.Description,
			}, nil
		}),
	); err != nil {
		return fmt.Errorf("copy movements: %w", err)
	}

	if len(classifications) > 0 {
		if _, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"classifications"},
			[]string{"movement_id", "kind", "confidence", "rule", "overridden_by", "overridden_at", "note"},
			pgx.CopyFromSlice(len(classifications), func(i int) ([]any, error) {
				c := classifications[i]
				return []any{
					c.MovementID, string(c.Kind), string(c.Confidence), string(c.Rule),
					c.OverriddenBy, c.OverriddenAt, c.Note,
				}, nil
			}),
		); err != nil {
			return fmt.Errorf("copy classifications: %w", err)
		}
	}
	return nil
}

func copyInvoices(ctx context.Context, tx pgx.Tx, ds *seed.Dataset) error {
	if len(ds.Invoices) > 0 {
		if _, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"invoices"},
			[]string{"id", "client_id", "type", "point_of_sale", "number", "issue_date", "currency", "net_cents", "vat_cents", "total_cents", "fx_rate_scaled", "receiver_name", "receiver_doc", "source", "created_at"},
			pgx.CopyFromSlice(len(ds.Invoices), func(i int) ([]any, error) {
				inv := ds.Invoices[i]
				var fx *int64
				if inv.FXRate != nil {
					v := int64(*inv.FXRate)
					fx = &v
				}
				return []any{
					inv.ID, inv.ClientID, string(inv.Type), inv.PointOfSale, inv.Number,
					inv.IssueDate, string(inv.Currency), inv.NetAmount.Amount, inv.VatAmount.Amount,
					inv.TotalAmount.Amount, fx, inv.ReceiverName, inv.ReceiverDoc, string(inv.Source), inv.CreatedAt,
				}, nil
			}),
		); err != nil {
			return fmt.Errorf("copy invoices: %w", err)
		}
	}

	if len(ds.PurchaseInvoices) > 0 {
		if _, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"purchase_invoices"},
			[]string{"id", "client_id", "issue_date", "supplier_name", "supplier_cuit", "net_cents", "vat_cents", "created_at"},
			pgx.CopyFromSlice(len(ds.PurchaseInvoices), func(i int) ([]any, error) {
				pi := ds.PurchaseInvoices[i]
				return []any{
					pi.ID, pi.ClientID, pi.IssueDate, pi.SupplierName, pi.SupplierCUIT,
					pi.NetAmount.Amount, pi.VatAmount.Amount, pi.CreatedAt,
				}, nil
			}),
		); err != nil {
			return fmt.Errorf("copy purchase_invoices: %w", err)
		}
	}
	return nil
}

func copyClientInputs(ctx context.Context, tx pgx.Tx, ds *seed.Dataset) error {
	if len(ds.ClientInputs) == 0 {
		return nil
	}
	if _, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"client_period_inputs"},
		[]string{"client_id", "period", "declared_none", "confirmed_at"},
		pgx.CopyFromSlice(len(ds.ClientInputs), func(i int) ([]any, error) {
			in := ds.ClientInputs[i]
			return []any{in.ClientID, in.Period.String(), in.DeclaredNoExternalIncome, in.ConfirmedAt}, nil
		}),
	); err != nil {
		return fmt.Errorf("copy client_period_inputs: %w", err)
	}

	type extRow struct {
		clientID string
		period   string
		seq      int
		amount   int64
		desc     string
	}
	var extRows []extRow
	for _, in := range ds.ClientInputs {
		for seq, item := range in.ExternalIncome {
			extRows = append(extRows, extRow{in.ClientID, in.Period.String(), seq, item.Amount.Amount, item.Description})
		}
	}
	if len(extRows) > 0 {
		if _, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"external_income_items"},
			[]string{"client_id", "period", "seq", "amount_cents", "description"},
			pgx.CopyFromSlice(len(extRows), func(i int) ([]any, error) {
				r := extRows[i]
				return []any{r.clientID, r.period, r.seq, r.amount, r.desc}, nil
			}),
		); err != nil {
			return fmt.Errorf("copy external_income_items: %w", err)
		}
	}
	return nil
}
