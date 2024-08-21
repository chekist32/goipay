-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS crypto_cache(
    coin coin_type PRIMARY KEY,
    last_synced_block_height BIGINT,
    synced_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('UTC', now())
);

INSERT INTO crypto_cache(coin) VALUES ('BTC');
INSERT INTO crypto_cache(coin) VALUES ('LTC');
INSERT INTO crypto_cache(coin) VALUES ('XMR');
INSERT INTO crypto_cache(coin) VALUES ('ETH');
INSERT INTO crypto_cache(coin) VALUES ('TON');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE crypto_cache CASCADE;
-- +goose StatementEnd
