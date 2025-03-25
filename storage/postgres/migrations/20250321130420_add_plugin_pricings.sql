-- +goose Up
-- +goose StatementBegin
CREATE TABLE plugin_pricings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    public_key TEXT NOT NULL,
    plugin_type plugin_type NOT NULL,
    is_ecdsa BOOLEAN DEFAULT TRUE,
    chain_code_hex TEXT NOT NULL,
    derive_path TEXT NOT NULL,
    signature TEXT NOT NULL,
    pricing JSONB NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugin_pricings;
-- +goose StatementEnd
