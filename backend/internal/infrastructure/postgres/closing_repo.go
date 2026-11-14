package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.ClosingRepository = (*ClosingRepository)(nil)
var _ ports.LatestExclusionDateReader = (*ClosingRepository)(nil)

// ClosingRepository implementa ports.ClosingRepository y LatestExclusionDateReader sobre PostgreSQL.
type ClosingRepository struct {
	pool *pgxpool.Pool
}

// NewClosingRepository crea un ClosingRepository en Postgres.
func NewClosingRepository(pool *pgxpool.Pool) *ClosingRepository {
	return &ClosingRepository{pool: pool}
}

type closingPayload struct {
	Mono           *tax.MonotributoAssessment `json:"mono,omitempty"`
	RI             *tax.RIAssessment          `json:"ri,omitempty"`
	Signals        []signal.Signal            `json:"signals,omitempty"`
	Reconciliation closing.Reconciliation     `json:"reconciliation"`
}

func extractDenorm(c closing.Closing) (
	primarySig string, primaryPrio int, regime, cat string,
	billedUSD, billedARS, acc12mARS int64, pctKCap int, toPay int64, toPayDue *time.Time,
) {
	primaryPrio = 999
	for _, s := range c.Signals {
		if s.Priority < primaryPrio || primarySig == "" {
			primaryPrio = s.Priority
			primarySig = string(s.Code)
		}
	}
	if c.Mono != nil {
		regime = string(client.RegimeMonotributo)
		cat = string(c.Mono.CurrentCategory)
		billedUSD = c.Mono.BilledInPeriodUSD.Amount
		billedARS = c.Mono.BilledInPeriodARS.Amount
		acc12mARS = c.Mono.Billed12mARS.Amount
		pctKCap = int(c.Mono.PctOfKCapBp)
		toPay = c.Mono.Fee.Amount
		d := c.Mono.FeeDueDate
		toPayDue = &d
	} else if c.RI != nil {
		regime = string(client.RegimeRI)
		billedUSD = c.RI.BilledInPeriodUSD.Amount
		billedARS = c.RI.BilledInPeriodARS.Amount
		toPay = c.RI.TotalToPay.Amount
		d := c.RI.VatDueDate
		if !c.RI.SelfEmployedDueDate.IsZero() && (d.IsZero() || c.RI.SelfEmployedDueDate.Before(d)) {
			d = c.RI.SelfEmployedDueDate
		}
		if !d.IsZero() {
			toPayDue = &d
		}
	}
	return
}

func scanClosing(db DBTX, ctx context.Context, row pgx.Row) (closing.Closing, error) {
	var c closing.Closing
	var perStr, statusStr, desc string
	var sev int16
	var payloadBytes, pubBytes []byte
	var pubAt, paidAt, firstPubAt *time.Time

	err := row.Scan(
		&c.ID, &c.ClientID, &perStr, &statusStr, &c.Version, &c.ResultHash,
		&sev, &desc, &payloadBytes, &pubBytes, &pubAt, &paidAt, &firstPubAt, &c.ComputedAt,
	)
	if err != nil {
		return closing.Closing{}, err
	}
	p, err := period.Parse(perStr)
	if err != nil {
		return closing.Closing{}, err
	}
	c.Period = p
	c.Status = closing.Status(statusStr)
	c.Severity = signal.Severity(sev)
	c.Description = desc
	c.ComputedAt = c.ComputedAt.In(period.Location())
	if paidAt != nil {
		t := paidAt.In(period.Location())
		c.PaidAt = &t
	}
	if firstPubAt != nil {
		t := firstPubAt.In(period.Location())
		c.FirstPublishedAt = &t
	}

	var pl closingPayload
	if err := json.Unmarshal(payloadBytes, &pl); err != nil {
		return closing.Closing{}, err
	}
	c.Mono = pl.Mono
	c.RI = pl.RI
	c.Signals = pl.Signals
	c.Reconciliation = pl.Reconciliation

	if len(pubBytes) > 0 {
		var pub closing.Published
		if err := json.Unmarshal(pubBytes, &pub); err != nil {
			return closing.Closing{}, err
		}
		c.Published = &pub
	}

	appRows, err := db.Query(ctx, `
		SELECT id, version, result_hash, approved_by, approved_at, client_note, batch
		FROM approvals WHERE closing_id = $1 ORDER BY approved_at ASC, id ASC
	`, c.ID)
	if err != nil {
		return closing.Closing{}, err
	}
	defer appRows.Close()
	for appRows.Next() {
		var a closing.Approval
		if err := appRows.Scan(&a.ID, &a.Version, &a.ResultHash, &a.ApprovedBy, &a.ApprovedAt, &a.Note, &a.Batch); err != nil {
			return closing.Closing{}, err
		}
		a.ApprovedAt = a.ApprovedAt.In(period.Location())
		c.Approvals = append(c.Approvals, a)
	}
	return c, appRows.Err()
}

func (r *ClosingRepository) Get(ctx context.Context, id string) (closing.Closing, error) {
	db := conn(ctx, r.pool)
	c, err := scanClosing(db, ctx, db.QueryRow(ctx, `
		SELECT id, client_id, period, status, version, result_hash, severity, description, payload, published, published_at, paid_at, first_published_at, computed_at
		FROM closings WHERE id = $1
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return closing.Closing{}, domain.ErrNotFound
		}
		return closing.Closing{}, err
	}
	return c, nil
}

func (r *ClosingRepository) ByClientPeriod(ctx context.Context, clientID string, p period.Period) (closing.Closing, bool, error) {
	db := conn(ctx, r.pool)
	c, err := scanClosing(db, ctx, db.QueryRow(ctx, `
		SELECT id, client_id, period, status, version, result_hash, severity, description, payload, published, published_at, paid_at, first_published_at, computed_at
		FROM closings WHERE client_id = $1 AND period = $2
	`, clientID, p.String()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return closing.Closing{}, false, nil
		}
		return closing.Closing{}, false, err
	}
	return c, true, nil
}

func (r *ClosingRepository) ListByPeriod(ctx context.Context, p period.Period) ([]closing.Closing, error) {
	db := conn(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, client_id, period, status, version, result_hash, severity, description, payload, published, published_at, paid_at, computed_at
		FROM closings WHERE period = $1 ORDER BY client_id ASC
	`, p.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []closing.Closing
	for rows.Next() {
		var c closing.Closing
		var perStr, statusStr, desc string
		var sev int16
		var payloadBytes, pubBytes []byte
		var pubAt, paidAt *time.Time
		if err := rows.Scan(&c.ID, &c.ClientID, &perStr, &statusStr, &c.Version, &c.ResultHash, &sev, &desc, &payloadBytes, &pubBytes, &pubAt, &paidAt, &c.ComputedAt); err != nil {
			return nil, err
		}
		c.Period = p
		c.Status = closing.Status(statusStr)
		c.Severity = signal.Severity(sev)
		c.Description = desc
		c.ComputedAt = c.ComputedAt.In(period.Location())
		if paidAt != nil {
			t := paidAt.In(period.Location())
			c.PaidAt = &t
		}
		var pl closingPayload
		if err := json.Unmarshal(payloadBytes, &pl); err != nil {
			return nil, err
		}
		c.Mono = pl.Mono
		c.RI = pl.RI
		c.Signals = pl.Signals
		c.Reconciliation = pl.Reconciliation
		if len(pubBytes) > 0 {
			var pub closing.Published
			if err := json.Unmarshal(pubBytes, &pub); err != nil {
				return nil, err
			}
			c.Published = &pub
		}
		ids = append(ids, c)
	}
	return ids, rows.Err()
}

func (r *ClosingRepository) ByClient(ctx context.Context, clientID string) ([]closing.Closing, error) {
	db := conn(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, client_id, period, status, version, result_hash, severity, description, payload, published, published_at, paid_at, computed_at
		FROM closings WHERE client_id = $1 ORDER BY period DESC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []closing.Closing
	for rows.Next() {
		var c closing.Closing
		var perStr, statusStr, desc string
		var sev int16
		var payloadBytes, pubBytes []byte
		var pubAt, paidAt *time.Time
		if err := rows.Scan(&c.ID, &c.ClientID, &perStr, &statusStr, &c.Version, &c.ResultHash, &sev, &desc, &payloadBytes, &pubBytes, &pubAt, &paidAt, &c.ComputedAt); err != nil {
			return nil, err
		}
		p, err := period.Parse(perStr)
		if err != nil {
			return nil, err
		}
		c.Period = p
		c.Status = closing.Status(statusStr)
		c.Severity = signal.Severity(sev)
		c.Description = desc
		c.ComputedAt = c.ComputedAt.In(period.Location())
		if paidAt != nil {
			t := paidAt.In(period.Location())
			c.PaidAt = &t
		}
		var pl closingPayload
		if err := json.Unmarshal(payloadBytes, &pl); err != nil {
			return nil, err
		}
		c.Mono = pl.Mono
		c.RI = pl.RI
		c.Signals = pl.Signals
		c.Reconciliation = pl.Reconciliation
		if len(pubBytes) > 0 {
			var pub closing.Published
			if err := json.Unmarshal(pubBytes, &pub); err != nil {
				return nil, err
			}
			c.Published = &pub
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ClosingRepository) Save(ctx context.Context, c closing.Closing) error {
	db := conn(ctx, r.pool)
	payloadBytes, pubBytes, pubAt, err := marshalClosing(c)
	if err != nil {
		return err
	}
	primarySig, primaryPrio, regime, cat, billedUSD, billedARS, acc12mARS, pctKCap, toPay, toPayDue := extractDenorm(c)

	_, err = db.Exec(ctx, `
		INSERT INTO closings (
			id, client_id, period, status, version, result_hash, severity,
			primary_signal, primary_priority, description, regime, category,
			billed_usd_cents, billed_ars_cents, acc12m_ars_cents, pct_k_cap_bp,
			to_pay_cents, to_pay_due, payload, published, published_at, paid_at, first_published_at, computed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24
		)
		ON CONFLICT (client_id, period) DO UPDATE SET
			id = EXCLUDED.id,
			status = EXCLUDED.status,
			version = EXCLUDED.version,
			result_hash = EXCLUDED.result_hash,
			severity = EXCLUDED.severity,
			primary_signal = EXCLUDED.primary_signal,
			primary_priority = EXCLUDED.primary_priority,
			description = EXCLUDED.description,
			regime = EXCLUDED.regime,
			category = EXCLUDED.category,
			billed_usd_cents = EXCLUDED.billed_usd_cents,
			billed_ars_cents = EXCLUDED.billed_ars_cents,
			acc12m_ars_cents = EXCLUDED.acc12m_ars_cents,
			pct_k_cap_bp = EXCLUDED.pct_k_cap_bp,
			to_pay_cents = EXCLUDED.to_pay_cents,
			to_pay_due = EXCLUDED.to_pay_due,
			payload = EXCLUDED.payload,
			published = EXCLUDED.published,
			published_at = EXCLUDED.published_at,
			paid_at = EXCLUDED.paid_at,
			first_published_at = COALESCE(closings.first_published_at, EXCLUDED.first_published_at),
			computed_at = EXCLUDED.computed_at
	`, c.ID, c.ClientID, c.Period.String(), string(c.Status), c.Version, c.ResultHash, int16(c.Severity),
		primarySig, int16(primaryPrio), c.Description, regime, cat,
		billedUSD, billedARS, acc12mARS, pctKCap,
		toPay, toPayDue, payloadBytes, pubBytes, pubAt, c.PaidAt, c.FirstPublishedAt, c.ComputedAt)
	if err != nil {
		return err
	}

	return r.syncChildren(ctx, db, c)
}

func (r *ClosingRepository) SaveIfVersion(ctx context.Context, c closing.Closing, expectedVersion int) error {
	db := conn(ctx, r.pool)
	payloadBytes, pubBytes, pubAt, err := marshalClosing(c)
	if err != nil {
		return err
	}
	primarySig, primaryPrio, regime, cat, billedUSD, billedARS, acc12mARS, pctKCap, toPay, toPayDue := extractDenorm(c)

	tag, err := db.Exec(ctx, `
		UPDATE closings SET
			status = $3,
			version = $4,
			result_hash = $5,
			severity = $6,
			primary_signal = $7,
			primary_priority = $8,
			description = $9,
			regime = $10,
			category = $11,
			billed_usd_cents = $12,
			billed_ars_cents = $13,
			acc12m_ars_cents = $14,
			pct_k_cap_bp = $15,
			to_pay_cents = $16,
			to_pay_due = $17,
			payload = $18,
			published = $19,
			published_at = $20,
			paid_at = $21,
			first_published_at = COALESCE(closings.first_published_at, $22),
			computed_at = $23
		WHERE id = $1 AND version = $2
		  AND NOT (status IN ('APPROVED', 'NOTIFIED') AND $3 IN ('APPROVED', 'NOTIFIED'))
	`, c.ID, expectedVersion, string(c.Status), c.Version, c.ResultHash, int16(c.Severity),
		primarySig, int16(primaryPrio), c.Description, regime, cat,
		billedUSD, billedARS, acc12mARS, pctKCap,
		toPay, toPayDue, payloadBytes, pubBytes, pubAt, c.PaidAt, c.FirstPublishedAt, c.ComputedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM closings WHERE id = $1)`, c.ID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return domain.ErrNotFound
		}
		return domain.ErrVersionConflict
	}

	return r.syncChildren(ctx, db, c)
}

func marshalClosing(c closing.Closing) ([]byte, []byte, *time.Time, error) {
	pl := closingPayload{
		Mono:           c.Mono,
		RI:             c.RI,
		Signals:        c.Signals,
		Reconciliation: c.Reconciliation,
	}
	payloadBytes, err := json.Marshal(pl)
	if err != nil {
		return nil, nil, nil, err
	}
	var pubBytes []byte
	var pubAt *time.Time
	if c.Published != nil {
		pubBytes, err = json.Marshal(c.Published)
		if err != nil {
			return nil, nil, nil, err
		}
		t := c.Published.PublishedAt
		pubAt = &t
	}
	return payloadBytes, pubBytes, pubAt, nil
}

func (r *ClosingRepository) syncChildren(ctx context.Context, db DBTX, c closing.Closing) error {
	if _, err := db.Exec(ctx, `DELETE FROM closing_signals WHERE closing_id = $1`, c.ID); err != nil {
		return err
	}
	for _, s := range c.Signals {
		if _, err := db.Exec(ctx, `
			INSERT INTO closing_signals (closing_id, code, severity)
			VALUES ($1, $2, $3)
			ON CONFLICT (closing_id, code) DO UPDATE SET severity = EXCLUDED.severity
		`, c.ID, string(s.Code), int16(s.Severity)); err != nil {
			return err
		}
	}
	for _, a := range c.Approvals {
		if _, err := db.Exec(ctx, `
			INSERT INTO approvals (id, closing_id, version, result_hash, approved_by, approved_at, client_note, batch)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, a.ID, c.ID, a.Version, a.ResultHash, a.ApprovedBy, a.ApprovedAt, a.Note, a.Batch); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClosingRepository) LatestExclusionDate(ctx context.Context, clientID string) (*time.Time, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT payload FROM closings WHERE client_id = $1
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var latest *time.Time
	for rows.Next() {
		var payloadBytes []byte
		if err := rows.Scan(&payloadBytes); err != nil {
			return nil, err
		}
		var pl closingPayload
		if err := json.Unmarshal(payloadBytes, &pl); err != nil {
			return nil, err
		}
		if pl.Mono != nil && pl.Mono.Excluded && pl.Mono.ExclusionDate != nil {
			if latest == nil || pl.Mono.ExclusionDate.After(*latest) {
				d := pl.Mono.ExclusionDate.In(period.Location())
				latest = &d
			}
		}
	}
	return latest, rows.Err()
}
