CREATE INDEX IF NOT EXISTS idx_closings_period_sev
  ON closings (period, severity DESC, primary_priority ASC, client_id ASC);

CREATE INDEX IF NOT EXISTS idx_closing_signals_code
  ON closing_signals (code);

CREATE INDEX IF NOT EXISTS idx_movements_client_occurred
  ON movements (client_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_invoices_client_issue
  ON invoices (client_id, issue_date);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
  ON outbox (sent_at) WHERE sent_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_clients_name_search
  ON clients (name_search text_pattern_ops);

