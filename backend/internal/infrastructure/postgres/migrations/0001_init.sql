CREATE TABLE IF NOT EXISTS clients (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  cuit TEXT NOT NULL UNIQUE,
  regime TEXT NOT NULL CHECK (regime IN ('MONOTRIBUTO','RI')),
  category TEXT NOT NULL DEFAULT '',
  ri_self_employed_cents BIGINT NOT NULL DEFAULT 0,
  ri_prev_vat_credit_cents BIGINT NOT NULL DEFAULT 0,
  name_search TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS regime_history (
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  seq INT NOT NULL,
  regime TEXT NOT NULL,
  valid_from TIMESTAMPTZ NOT NULL,
  valid_to TIMESTAMPTZ,
  reason TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (client_id, seq)
);

CREATE TABLE IF NOT EXISTS income_tax_advances (
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  period TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  due_date TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (client_id, period)
);

CREATE TABLE IF NOT EXISTS movements (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  currency TEXT NOT NULL,
  fx_rate_scaled BIGINT NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  cp_name TEXT NOT NULL DEFAULT '',
  cp_country TEXT NOT NULL DEFAULT '',
  cp_kind TEXT NOT NULL DEFAULT '',
  cp_own BOOLEAN NOT NULL DEFAULT false,
  description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS classifications (
  movement_id TEXT PRIMARY KEY REFERENCES movements(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  confidence TEXT NOT NULL,
  rule TEXT NOT NULL,
  overridden_by TEXT NOT NULL DEFAULT '',
  overridden_at TIMESTAMPTZ,
  note TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS invoices (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  point_of_sale INT NOT NULL,
  number BIGINT NOT NULL,
  issue_date TIMESTAMPTZ NOT NULL,
  currency TEXT NOT NULL,
  net_cents BIGINT NOT NULL,
  vat_cents BIGINT NOT NULL,
  total_cents BIGINT NOT NULL,
  fx_rate_scaled BIGINT,
  receiver_name TEXT NOT NULL DEFAULT '',
  receiver_doc TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (client_id, type, point_of_sale, number)
);

CREATE TABLE IF NOT EXISTS purchase_invoices (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  issue_date TIMESTAMPTZ NOT NULL,
  supplier_name TEXT NOT NULL DEFAULT '',
  supplier_cuit TEXT NOT NULL DEFAULT '',
  net_cents BIGINT NOT NULL,
  vat_cents BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS client_period_inputs (
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  period TEXT NOT NULL,
  declared_none BOOLEAN NOT NULL DEFAULT false,
  confirmed_at TIMESTAMPTZ,
  PRIMARY KEY (client_id, period)
);

CREATE TABLE IF NOT EXISTS external_income_items (
  client_id TEXT NOT NULL,
  period TEXT NOT NULL,
  seq INT NOT NULL,
  amount_cents BIGINT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (client_id, period, seq),
  FOREIGN KEY (client_id, period) REFERENCES client_period_inputs(client_id, period) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS closings (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  period TEXT NOT NULL,
  status TEXT NOT NULL,
  version INT NOT NULL,
  result_hash TEXT NOT NULL,
  severity SMALLINT NOT NULL,
  primary_signal TEXT NOT NULL DEFAULT '',
  primary_priority SMALLINT NOT NULL DEFAULT 999,
  description TEXT NOT NULL DEFAULT '',
  regime TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  billed_usd_cents BIGINT NOT NULL DEFAULT 0,
  billed_ars_cents BIGINT NOT NULL DEFAULT 0,
  acc12m_ars_cents BIGINT NOT NULL DEFAULT 0,
  pct_k_cap_bp INT NOT NULL DEFAULT 0,
  to_pay_cents BIGINT NOT NULL DEFAULT 0,
  to_pay_due TIMESTAMPTZ,
  payload JSONB NOT NULL,
  published JSONB,
  published_at TIMESTAMPTZ,
  paid_at TIMESTAMPTZ,
  computed_at TIMESTAMPTZ NOT NULL,
  UNIQUE (client_id, period)
);

CREATE TABLE IF NOT EXISTS closing_signals (
  closing_id TEXT NOT NULL REFERENCES closings(id) ON DELETE CASCADE,
  code TEXT NOT NULL,
  severity SMALLINT NOT NULL,
  PRIMARY KEY (closing_id, code)
);

CREATE TABLE IF NOT EXISTS approvals (
  id TEXT PRIMARY KEY,
  closing_id TEXT NOT NULL REFERENCES closings(id) ON DELETE CASCADE,
  version INT NOT NULL,
  result_hash TEXT NOT NULL,
  approved_by TEXT NOT NULL,
  approved_at TIMESTAMPTZ NOT NULL,
  client_note TEXT NOT NULL DEFAULT '',
  batch BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS notifications (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  closing_id TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL,
  dedup_key TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  read_at TIMESTAMPTZ,
  UNIQUE (client_id, closing_id, kind, dedup_key)
);

CREATE TABLE IF NOT EXISTS outbox (
  id TEXT PRIMARY KEY,
  channel TEXT NOT NULL,
  client_id TEXT NOT NULL,
  subject TEXT NOT NULL,
  body TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  sent_at TIMESTAMPTZ,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS parameters (
  id SMALLINT PRIMARY KEY CHECK (id = 1),
  data JSONB NOT NULL,
  updated_at TIMESTAMPTZ,
  updated_by TEXT NOT NULL DEFAULT ''
);

