-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS xmr_crypto_data(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    priv_view_key TEXT NOT NULL UNIQUE,
    pub_spend_key TEXT NOT NULL UNIQUE,
    last_major_index INTEGER NOT NULL DEFAULT 0,
    last_minor_index INTEGER NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE xmr_crypto_data CASCADE;
-- +goose StatementEnd
