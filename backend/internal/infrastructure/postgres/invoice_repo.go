package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.InvoiceRepository = (*InvoiceRepository)(nil)

// InvoiceRepository implementa ports.InvoiceRepository sobre PostgreSQL.
type InvoiceRepository struct {
	pool *pgxpool.Pool
}

// NewInvoiceRepository crea un InvoiceRepository en Postgres.
func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{pool: pool}
}

func scanInvoice(row pgx.Row) (invoice.Invoice, error) {
	var inv invoice.Invoice
	var typ, cur, src string
	var netC, vatC, totC int64
	var fxScaled *int64
	err := row.Scan(
		&inv.ID, &inv.ClientID, &typ, &inv.PointOfSale, &inv.Number, &inv.IssueDate,
		&cur, &netC, &vatC, &totC, &fxScaled, &inv.ReceiverName, &inv.ReceiverDoc, &src, &inv.CreatedAt,
	)
	if err != nil {
		return invoice.Invoice{}, err
	}
	inv.Type = invoice.InvoiceType(typ)
	inv.Currency = money.Currency(cur)
	inv.NetAmount = money.New(netC, inv.Currency)
	inv.VatAmount = money.New(vatC, inv.Currency)
	inv.TotalAmount = money.New(totC, inv.Currency)
	if fxScaled != nil {
		r := money.Rate(*fxScaled)
		inv.FXRate = &r
	}
	inv.Source = invoice.InvoiceSource(src)
	inv.IssueDate = inv.IssueDate.In(period.Location())
	inv.CreatedAt = inv.CreatedAt.In(period.Location())
	return inv, nil
}

func (r *InvoiceRepository) ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]invoice.Invoice, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, type, point_of_sale, number, issue_date, currency, net_cents, vat_cents, total_cents, fx_rate_scaled, receiver_name, receiver_doc, source, created_at
		FROM invoices
		WHERE client_id = $1 AND issue_date >= $2 AND issue_date <= $3
		ORDER BY issue_date ASC, id ASC
	`, clientID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []invoice.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *InvoiceRepository) ByClient(ctx context.Context, clientID string) ([]invoice.Invoice, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, type, point_of_sale, number, issue_date, currency, net_cents, vat_cents, total_cents, fx_rate_scaled, receiver_name, receiver_doc, source, created_at
		FROM invoices
		WHERE client_id = $1
		ORDER BY issue_date ASC, id ASC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []invoice.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *InvoiceRepository) Exists(ctx context.Context, clientID string, key invoice.Key) (bool, error) {
	var exists bool
	err := conn(ctx, r.pool).QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM invoices WHERE client_id = $1 AND type = $2 AND point_of_sale = $3 AND number = $4
		)
	`, clientID, string(key.Type), key.PointOfSale, key.Number).Scan(&exists)
	return exists, err
}

func (r *InvoiceRepository) Get(ctx context.Context, id string) (invoice.Invoice, error) {
	inv, err := scanInvoice(conn(ctx, r.pool).QueryRow(ctx, `
		SELECT id, client_id, type, point_of_sale, number, issue_date, currency, net_cents, vat_cents, total_cents, fx_rate_scaled, receiver_name, receiver_doc, source, created_at
		FROM invoices WHERE id = $1
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoice.Invoice{}, domain.ErrNotFound
		}
		return invoice.Invoice{}, err
	}
	return inv, nil
}

func (r *InvoiceRepository) Save(ctx context.Context, inv invoice.Invoice) error {
	var fxScaled *int64
	if inv.FXRate != nil {
		v := int64(*inv.FXRate)
		fxScaled = &v
	}
	_, err := conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO invoices (id, client_id, type, point_of_sale, number, issue_date, currency, net_cents, vat_cents, total_cents, fx_rate_scaled, receiver_name, receiver_doc, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			issue_date = EXCLUDED.issue_date,
			currency = EXCLUDED.currency,
			net_cents = EXCLUDED.net_cents,
			vat_cents = EXCLUDED.vat_cents,
			total_cents = EXCLUDED.total_cents,
			fx_rate_scaled = EXCLUDED.fx_rate_scaled,
			receiver_name = EXCLUDED.receiver_name,
			receiver_doc = EXCLUDED.receiver_doc
	`, inv.ID, inv.ClientID, string(inv.Type), inv.PointOfSale, inv.Number, inv.IssueDate,
		string(inv.Currency), inv.NetAmount.Amount, inv.VatAmount.Amount, inv.TotalAmount.Amount,
		fxScaled, inv.ReceiverName, inv.ReceiverDoc, string(inv.Source), inv.CreatedAt)
	return err
}

func (r *InvoiceRepository) SaveAll(ctx context.Context, invs []invoice.Invoice) error {
	for _, inv := range invs {
		if err := r.Save(ctx, inv); err != nil {
			return err
		}
	}
	return nil
}
