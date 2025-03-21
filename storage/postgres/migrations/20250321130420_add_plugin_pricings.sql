-- +goose Up
-- +goose StatementBegin
CREATE TABLE plugin_pricings (
    id UUID PRIMARY KEY,
    public_key TEXT NOT NULL,
    plugin_type plugin_type NOT NULL,
    signature TEXT NOT NULL,
    pricing JSONB NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugin_pricings;
-- +goose StatementEnd
