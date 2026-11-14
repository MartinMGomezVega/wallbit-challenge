package postgres

import (
	"context"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.PurchaseInvoiceRepository = (*PurchaseInvoiceRepository)(nil)

// PurchaseInvoiceRepository implementa ports.PurchaseInvoiceRepository sobre PostgreSQL.
type PurchaseInvoiceRepository struct {
	pool *pgxpool.Pool
}

// NewPurchaseInvoiceRepository crea un PurchaseInvoiceRepository en Postgres.
func NewPurchaseInvoiceRepository(pool *pgxpool.Pool) *PurchaseInvoiceRepository {
	return &PurchaseInvoiceRepository{pool: pool}
}

func scanPurchaseInvoice(row pgx.Row) (invoice.PurchaseInvoice, error) {
	var pi invoice.PurchaseInvoice
	var netC, vatC int64
	err := row.Scan(&pi.ID, &pi.ClientID, &pi.IssueDate, &pi.SupplierName, &pi.SupplierCUIT, &netC, &vatC, &pi.CreatedAt)
	if err != nil {
		return invoice.PurchaseInvoice{}, err
	}
	pi.NetAmount = money.New(netC, money.ARS)
	pi.VatAmount = money.New(vatC, money.ARS)
	pi.IssueDate = pi.IssueDate.In(period.Location())
	pi.CreatedAt = pi.CreatedAt.In(period.Location())
	return pi, nil
}

func (r *PurchaseInvoiceRepository) ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]invoice.PurchaseInvoice, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, issue_date, supplier_name, supplier_cuit, net_cents, vat_cents, created_at
		FROM purchase_invoices
		WHERE client_id = $1 AND issue_date >= $2 AND issue_date <= $3
		ORDER BY issue_date ASC, id ASC
	`, clientID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []invoice.PurchaseInvoice
	for rows.Next() {
		pi, err := scanPurchaseInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, pi)
	}
	return out, rows.Err()
}

func (r *PurchaseInvoiceRepository) ByClient(ctx context.Context, clientID string) ([]invoice.PurchaseInvoice, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, issue_date, supplier_name, supplier_cuit, net_cents, vat_cents, created_at
		FROM purchase_invoices
		WHERE client_id = $1
		ORDER BY issue_date ASC, id ASC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []invoice.PurchaseInvoice
	for rows.Next() {
		pi, err := scanPurchaseInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, pi)
	}
	return out, rows.Err()
}

func (r *PurchaseInvoiceRepository) Save(ctx context.Context, pi invoice.PurchaseInvoice) error {
	_, err := conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO purchase_invoices (id, client_id, issue_date, supplier_name, supplier_cuit, net_cents, vat_cents, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			issue_date = EXCLUDED.issue_date,
			supplier_name = EXCLUDED.supplier_name,
			supplier_cuit = EXCLUDED.supplier_cuit,
			net_cents = EXCLUDED.net_cents,
			vat_cents = EXCLUDED.vat_cents
	`, pi.ID, pi.ClientID, pi.IssueDate, pi.SupplierName, pi.SupplierCUIT, pi.NetAmount.Amount, pi.VatAmount.Amount, pi.CreatedAt)
	return err
}

func (r *PurchaseInvoiceRepository) SaveAll(ctx context.Context, pis []invoice.PurchaseInvoice) error {
	for _, pi := range pis {
		if err := r.Save(ctx, pi); err != nil {
			return err
		}
	}
	return nil
}
