package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/internal/types"
)

const PLUGINS_TABLE = "plugins"

func (p *PostgresBackend) FindPluginById(ctx context.Context, id string) (*types.Plugin, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, PLUGINS_TABLE)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	plugin, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.Plugin])
	if err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (p *PostgresBackend) FindPlugins(ctx context.Context) ([]types.Plugin, error) {
	if p.pool == nil {
		return []types.Plugin{}, fmt.Errorf("database pool is nil")
	}

	query := fmt.Sprintf(`SELECT * FROM %s`, PLUGINS_TABLE)

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	plugins, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Plugin])
	if err != nil {
		return nil, err
	}

	return plugins, nil
}

func (p *PostgresBackend) CreatePlugin(ctx context.Context, pluginDto types.PluginCreateDto) (*types.Plugin, error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		title,
		description,
		metadata,
		server_endpoint,
		vaults,
		pricing_id
	) VALUES (
		@Title,
		@Description,
		@Metadata,
		@ServerEndpoint,
		@Vaults,
		@PricingID
	) RETURNING id;`, PLUGINS_TABLE)
	args := pgx.NamedArgs{
		"Title":          pluginDto.Title,
		"Description":    pluginDto.Description,
		"Metadata":       pluginDto.Metadata,
		"ServerEndpoint": pluginDto.ServerEndpoint,
		"Vaults":         pluginDto.Vaults,
		"PricingID":      pluginDto.PricingID,
	}

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return nil, err
	}

	return p.FindPluginById(ctx, createdId)
}
