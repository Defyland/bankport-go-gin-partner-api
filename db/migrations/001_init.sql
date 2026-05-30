CREATE TABLE partners (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE developer_apps (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  name TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'suspended', 'deleted')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
  id TEXT PRIMARY KEY,
  developer_app_id TEXT NOT NULL REFERENCES developer_apps(id),
  key_hash TEXT NOT NULL UNIQUE,
  scopes TEXT[] NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'revoked')),
  rotated_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  currency CHAR(3) NOT NULL,
  available_balance_cents BIGINT NOT NULL CHECK (available_balance_cents >= 0),
  pending_balance_cents BIGINT NOT NULL CHECK (pending_balance_cents >= 0),
  version BIGINT NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_partner_id ON accounts(partner_id);

CREATE TABLE statement_entries (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL REFERENCES accounts(id),
  entry_type TEXT NOT NULL CHECK (entry_type IN ('credit', 'debit')),
  description TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_statement_entries_account_time ON statement_entries(account_id, occurred_at DESC);

CREATE TABLE idempotency_keys (
  partner_id TEXT NOT NULL REFERENCES partners(id),
  route TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  request_hash TEXT NOT NULL,
  response_status INTEGER NOT NULL,
  response_body JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (partner_id, route, idempotency_key)
);

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

CREATE TABLE pix_transfers (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  source_account_id TEXT NOT NULL REFERENCES accounts(id),
  amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
  refunded_amount_cents BIGINT NOT NULL DEFAULT 0 CHECK (refunded_amount_cents >= 0 AND refunded_amount_cents <= amount_cents),
  currency CHAR(3) NOT NULL,
  pix_key TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('accepted', 'rejected', 'settled')),
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pix_transfers_partner_time ON pix_transfers(partner_id, created_at DESC);

-- Production refund acceptance should update pix_transfers with a guard:
-- refunded_amount_cents = refunded_amount_cents + :amount
-- WHERE refunded_amount_cents + :amount <= amount_cents
-- This prevents cumulative partial refunds from exceeding the original transfer.

CREATE TABLE payouts (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  account_id TEXT NOT NULL REFERENCES accounts(id),
  amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
  currency CHAR(3) NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('queued', 'processing', 'paid', 'failed')),
  bank_code TEXT NOT NULL,
  branch TEXT NOT NULL,
  account_number TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payouts_partner_time ON payouts(partner_id, created_at DESC);

CREATE TABLE refunds (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  account_id TEXT NOT NULL REFERENCES accounts(id),
  original_transaction_id TEXT NOT NULL,
  amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
  currency CHAR(3) NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('accepted', 'rejected', 'settled')),
  reason TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refunds_original_transaction ON refunds(partner_id, original_transaction_id);

CREATE TABLE webhook_endpoints (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  url TEXT NOT NULL,
  event_types TEXT[] NOT NULL,
  secret_id TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'disabled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_webhook_endpoints_partner ON webhook_endpoints(partner_id);

CREATE TABLE webhook_deliveries (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  endpoint_id TEXT NOT NULL REFERENCES webhook_endpoints(id),
  event_id TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('queued', 'delivered', 'failed', 'dead_lettered')),
  signature TEXT NOT NULL,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_webhook_deliveries_retry ON webhook_deliveries(status, next_attempt_at);

CREATE TABLE outbox_events (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  event_type TEXT NOT NULL,
  schema_version TEXT NOT NULL,
  payload JSONB NOT NULL,
  correlation_id TEXT NOT NULL,
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_events_unpublished ON outbox_events(created_at) WHERE published_at IS NULL;

CREATE TABLE audit_entries (
  id TEXT PRIMARY KEY,
  partner_id TEXT NOT NULL REFERENCES partners(id),
  request_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  action TEXT NOT NULL,
  resource_id TEXT,
  status TEXT NOT NULL,
  reason TEXT,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_entries_partner_time ON audit_entries(partner_id, occurred_at DESC);
