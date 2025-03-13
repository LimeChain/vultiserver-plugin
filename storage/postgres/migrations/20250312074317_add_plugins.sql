-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB NOT NULL,
    server_endpoint VARCHAR(255) NOT NULL,
    vaults VARCHAR(255) NOT NULL,
    pricing_id INT NOT NULL,
    CONSTRAINT fk_pricing FOREIGN KEY (pricing_id) REFERENCES pricings(id) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS plugins;
-- +goose StatementEnd
