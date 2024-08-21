-- +goose Up
-- +goose StatementBegin
CREATE TYPE invoice_status_type AS ENUM (
  'PENDING',
  'PENDING_MEMPOOL',
  'EXPIRED',
  'CONFIRMED'
);

CREATE TABLE IF NOT EXISTS invoices(
    id UUID	PRIMARY KEY DEFAULT gen_random_uuid(),
    crypto_address TEXT NOT NULL,
    coin coin_type NOT NULL,
    required_amount DOUBLE PRECISION NOT NULL,
    actual_amount DOUBLE PRECISION,
    confirmations_required SMALLINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('UTC', now()),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    status invoice_status_type NOT NULL DEFAULT 'PENDING',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    tx_id TEXT,
    user_id UUID NOT NULL REFERENCES users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE invoice_status_type CASCADE;

DROP TABLE invoices CASCADE;
-- +goose StatementEnd