-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS crypto_data(
    user_id UUID PRIMARY KEY REFERENCES users (id),
    xmr_id UUID REFERENCES xmr_crypto_data (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE crypto_data CASCADE;
-- +goose StatementEnd
