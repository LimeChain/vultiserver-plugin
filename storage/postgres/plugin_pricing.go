package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/internal/types"
)

const PLUGIN_PRICINGS_TABLE = "plugin_pricings"

func (p *PostgresBackend) findPluginPricingById(ctx context.Context, id string) (*types.PluginPricing, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, PLUGIN_PRICINGS_TABLE)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	plugin, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.PluginPricing])
	if err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (p *PostgresBackend) CreatePluginPricing(
	ctx context.Context,
	pluginPricingDto types.PluginPricingCreateDto,
) (*types.PluginPricing, error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		public_key,
		plugin_type,
		is_ecdsa,
		chain_code_hex,
		derive_path,
		signature,
		pricing
	) VALUES (
		@PublicKey,
		@PluginType,
		@IsEcdsa,
		@ChainCodeHex,
		@DerivePath,
		@Signature,
		@Pricing
	) RETURNING id;`, PLUGIN_PRICINGS_TABLE)
	args := pgx.NamedArgs{
		"PublicKey":    pluginPricingDto.PublicKey,
		"PluginType":   pluginPricingDto.PluginType,
		"IsEcdsa":      pluginPricingDto.IsEcdsa,
		"ChainCodeHex": pluginPricingDto.ChainCodeHex,
		"DerivePath":   pluginPricingDto.DerivePath,
		"Signature":    pluginPricingDto.Signature,
		"Pricing":      pluginPricingDto.Pricing,
	}

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return nil, err
	}

	return p.findPluginPricingById(ctx, createdId)
}
