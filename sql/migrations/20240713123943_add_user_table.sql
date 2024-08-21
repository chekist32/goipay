-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users(
	id UUID	PRIMARY KEY DEFAULT gen_random_uuid()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users CASCADE;
-- +goose StatementEnd
