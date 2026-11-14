package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.ClientRepository = (*ClientRepository)(nil)

// ClientRepository implementa ports.ClientRepository sobre PostgreSQL.
type ClientRepository struct {
	pool *pgxpool.Pool
}

// NewClientRepository crea un ClientRepository en Postgres.
func NewClientRepository(pool *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{pool: pool}
}

func (r *ClientRepository) List(ctx context.Context) ([]client.Client, error) {
	db := conn(ctx, r.pool)
	rows, err := db.Query(ctx, `
		SELECT id, name, email, cuit, regime, category, ri_self_employed_cents, ri_prev_vat_credit_cents, created_at
		FROM clients
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []client.Client
	byID := make(map[string]int)
	for rows.Next() {
		var c client.Client
		var cuitStr, regimeStr, catStr string
		var seCents, prevVatCents int64
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &cuitStr, &regimeStr, &catStr, &seCents, &prevVatCents, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CUIT = client.CUIT(cuitStr)
		c.Regime = client.Regime(regimeStr)
		c.MonotributoCategory = tax.Category(catStr)
		c.CreatedAt = c.CreatedAt.In(period.Location())
		if c.Regime == client.RegimeRI || seCents != 0 || prevVatCents != 0 {
			c.RIProfile = &client.RIProfile{
				SelfEmployedMonthlyContribution: money.New(seCents, money.ARS),
				PreviousVatCredit:               money.New(prevVatCents, money.ARS),
			}
		}
		byID[c.ID] = len(out)
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return out, nil
	}

	histRows, err := db.Query(ctx, `
		SELECT client_id, regime, valid_from, valid_to, reason, note
		FROM regime_history
		ORDER BY client_id ASC, seq ASC
	`)
	if err != nil {
		return nil, err
	}
	defer histRows.Close()
	for histRows.Next() {
		var cid, regStr, reason, note string
		var from time.Time
		var to *time.Time
		if err := histRows.Scan(&cid, &regStr, &from, &to, &reason, &note); err != nil {
			return nil, err
		}
		if idx, ok := byID[cid]; ok {
			var toLoc *time.Time
			if to != nil {
				t := to.In(period.Location())
				toLoc = &t
			}
			out[idx].RegimeHistory = append(out[idx].RegimeHistory, client.RegimeHistoryItem{
				Regime: client.Regime(regStr),
				From:   from.In(period.Location()),
				To:     toLoc,
				Reason: reason,
				Note:   note,
			})
		}
	}
	if err := histRows.Err(); err != nil {
		return nil, err
	}

	advRows, err := db.Query(ctx, `
		SELECT client_id, period, amount_cents, due_date
		FROM income_tax_advances
	`)
	if err != nil {
		return nil, err
	}
	defer advRows.Close()
	for advRows.Next() {
		var cid, per string
		var amt int64
		var due time.Time
		if err := advRows.Scan(&cid, &per, &amt, &due); err != nil {
			return nil, err
		}
		if idx, ok := byID[cid]; ok {
			p, err := period.Parse(per)
			if err != nil {
				return nil, err
			}
			if out[idx].RIProfile == nil {
				out[idx].RIProfile = &client.RIProfile{}
			}
			out[idx].RIProfile.IncomeTaxAdvances = append(out[idx].RIProfile.IncomeTaxAdvances, client.IncomeTaxAdvance{
				Period:  p,
				Amount:  money.New(amt, money.ARS),
				DueDate: due.In(period.Location()),
			})
		}
	}
	return out, advRows.Err()
}

func (r *ClientRepository) Get(ctx context.Context, id string) (client.Client, error) {
	db := conn(ctx, r.pool)
	var c client.Client
	var cuitStr, regimeStr, catStr string
	var seCents, prevVatCents int64
	err := db.QueryRow(ctx, `
		SELECT id, name, email, cuit, regime, category, ri_self_employed_cents, ri_prev_vat_credit_cents, created_at
		FROM clients WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Email, &cuitStr, &regimeStr, &catStr, &seCents, &prevVatCents, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return client.Client{}, domain.ErrNotFound
		}
		return client.Client{}, err
	}
	c.CUIT = client.CUIT(cuitStr)
	c.Regime = client.Regime(regimeStr)
	c.MonotributoCategory = tax.Category(catStr)
	c.CreatedAt = c.CreatedAt.In(period.Location())
	if c.Regime == client.RegimeRI || seCents != 0 || prevVatCents != 0 {
		c.RIProfile = &client.RIProfile{
			SelfEmployedMonthlyContribution: money.New(seCents, money.ARS),
			PreviousVatCredit:               money.New(prevVatCents, money.ARS),
		}
	}

	histRows, err := db.Query(ctx, `
		SELECT regime, valid_from, valid_to, reason, note
		FROM regime_history WHERE client_id = $1 ORDER BY seq ASC
	`, id)
	if err != nil {
		return client.Client{}, err
	}
	defer histRows.Close()
	for histRows.Next() {
		var regStr, reason, note string
		var from time.Time
		var to *time.Time
		if err := histRows.Scan(&regStr, &from, &to, &reason, &note); err != nil {
			return client.Client{}, err
		}
		var toLoc *time.Time
		if to != nil {
			t := to.In(period.Location())
			toLoc = &t
		}
		c.RegimeHistory = append(c.RegimeHistory, client.RegimeHistoryItem{
			Regime: client.Regime(regStr),
			From:   from.In(period.Location()),
			To:     toLoc,
			Reason: reason,
			Note:   note,
		})
	}
	if err := histRows.Err(); err != nil {
		return client.Client{}, err
	}

	advRows, err := db.Query(ctx, `
		SELECT period, amount_cents, due_date
		FROM income_tax_advances WHERE client_id = $1
	`, id)
	if err != nil {
		return client.Client{}, err
	}
	defer advRows.Close()
	for advRows.Next() {
		var per string
		var amt int64
		var due time.Time
		if err := advRows.Scan(&per, &amt, &due); err != nil {
			return client.Client{}, err
		}
		p, err := period.Parse(per)
		if err != nil {
			return client.Client{}, err
		}
		if c.RIProfile == nil {
			c.RIProfile = &client.RIProfile{}
		}
		c.RIProfile.IncomeTaxAdvances = append(c.RIProfile.IncomeTaxAdvances, client.IncomeTaxAdvance{
			Period:  p,
			Amount:  money.New(amt, money.ARS),
			DueDate: due.In(period.Location()),
		})
	}
	return c, advRows.Err()
}

func (r *ClientRepository) Save(ctx context.Context, c client.Client) error {
	return r.saveOne(ctx, conn(ctx, r.pool), c)
}

func (r *ClientRepository) SaveAll(ctx context.Context, cs []client.Client) error {
	db := conn(ctx, r.pool)
	for _, c := range cs {
		if err := r.saveOne(ctx, db, c); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientRepository) saveOne(ctx context.Context, db DBTX, c client.Client) error {
	var seCents, prevVatCents int64
	if c.RIProfile != nil {
		seCents = c.RIProfile.SelfEmployedMonthlyContribution.Amount
		prevVatCents = c.RIProfile.PreviousVatCredit.Amount
	}
	_, err := db.Exec(ctx, `
		INSERT INTO clients (id, name, email, cuit, regime, category, ri_self_employed_cents, ri_prev_vat_credit_cents, name_search, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			cuit = EXCLUDED.cuit,
			regime = EXCLUDED.regime,
			category = EXCLUDED.category,
			ri_self_employed_cents = EXCLUDED.ri_self_employed_cents,
			ri_prev_vat_credit_cents = EXCLUDED.ri_prev_vat_credit_cents,
			name_search = EXCLUDED.name_search
	`, c.ID, c.Name, c.Email, string(c.CUIT), string(c.Regime), string(c.MonotributoCategory), seCents, prevVatCents, strings.ToLower(c.Name), c.CreatedAt)
	if err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `DELETE FROM regime_history WHERE client_id = $1`, c.ID); err != nil {
		return err
	}
	for seq, h := range c.RegimeHistory {
		if _, err := db.Exec(ctx, `
			INSERT INTO regime_history (client_id, seq, regime, valid_from, valid_to, reason, note)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, c.ID, seq, string(h.Regime), h.From, h.To, h.Reason, h.Note); err != nil {
			return err
		}
	}

	if _, err := db.Exec(ctx, `DELETE FROM income_tax_advances WHERE client_id = $1`, c.ID); err != nil {
		return err
	}
	if c.RIProfile != nil {
		for _, adv := range c.RIProfile.IncomeTaxAdvances {
			if _, err := db.Exec(ctx, `
				INSERT INTO income_tax_advances (client_id, period, amount_cents, due_date)
				VALUES ($1, $2, $3, $4)
			`, c.ID, adv.Period.String(), adv.Amount.Amount, adv.DueDate); err != nil {
				return err
			}
		}
	}
	return nil
}
