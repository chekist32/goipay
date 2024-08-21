-- +goose Up
-- +goose StatementBegin
CREATE TYPE coin_type AS ENUM (
  'XMR',
  'BTC',
  'LTC',
  'ETH',
  'TON'
);

CREATE TABLE IF NOT EXISTS crypto_addresses(
    id UUID	PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT NOT NULL UNIQUE,
    coin coin_type NOT NULL,
    is_occupied BOOLEAN NOT NULL DEFAULT FALSE,
    user_id UUID NOT NULL REFERENCES users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE coin_type CASCADE;

DROP TABLE crypto_addresses CASCADE;
-- +goose StatementEnd
