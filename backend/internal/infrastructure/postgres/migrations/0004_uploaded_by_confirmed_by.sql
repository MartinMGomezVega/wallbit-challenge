ALTER TABLE invoices
  ADD COLUMN IF NOT EXISTS uploaded_by TEXT NOT NULL DEFAULT 'CLIENT';

ALTER TABLE client_period_inputs
  ADD COLUMN IF NOT EXISTS confirmed_by TEXT,
  ADD COLUMN IF NOT EXISTS confirmation_note TEXT;

