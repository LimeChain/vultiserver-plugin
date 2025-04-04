package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/types"
)

const PLUGINS_TABLE = "plugins"

func (p *PostgresBackend) FindPluginById(ctx context.Context, id string) (types.PluginDto, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, PLUGINS_TABLE)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return types.PluginDto{}, err
	}

	plugin, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.Plugin])
	if err != nil {
		return types.PluginDto{}, err
	}

	ratings, err := p.FindRatingByPluginId(ctx, plugin.ID)
	if err != nil {
		return types.PluginDto{}, err
	}

	var pluginDto types.PluginDto

	pluginDto.ID = plugin.ID
	pluginDto.CreatedAt = plugin.CreatedAt
	pluginDto.UpdatedAt = plugin.UpdatedAt
	pluginDto.Type = plugin.Type
	pluginDto.Title = plugin.Title
	pluginDto.Description = plugin.Description
	pluginDto.Metadata = plugin.Metadata
	pluginDto.ServerEndpoint = plugin.ServerEndpoint
	pluginDto.Ratings = ratings

	return pluginDto, nil
}

func (p *PostgresBackend) FindPlugins(ctx context.Context, skip int, take int, sort string) (types.PlugisDto, error) {
	if p.pool == nil {
		return types.PlugisDto{}, fmt.Errorf("database pool is nil")
	}

	allowedSortingColumns := map[string]bool{"updated_at": true, "created_at": true, "title": true}
	orderBy, orderDirection := common.GetSortingCondition(sort, allowedSortingColumns)

	query := fmt.Sprintf(`
		SELECT *, COUNT(*) OVER() AS total_count
		FROM %s 
		ORDER BY %s %s
		LIMIT $1 OFFSET $2`, PLUGINS_TABLE, orderBy, orderDirection)

	rows, err := p.pool.Query(ctx, query, take, skip)
	if err != nil {
		return types.PlugisDto{}, err
	}

	defer rows.Close()

	var plugins []types.Plugin
	var totalCount int

	for rows.Next() {
		var plugin types.Plugin

		err := rows.Scan(
			&plugin.ID,
			&plugin.CreatedAt,
			&plugin.UpdatedAt,
			&plugin.Type,
			&plugin.Title,
			&plugin.Description,
			&plugin.Metadata,
			&plugin.ServerEndpoint,
			&plugin.PricingID,
			&totalCount,
		)
		if err != nil {
			return types.PlugisDto{}, err
		}

		plugins = append(plugins, plugin)
	}

	pluginsDto := types.PlugisDto{
		Plugins:    plugins,
		TotalCount: totalCount,
	}

	return pluginsDto, nil
}

func (p *PostgresBackend) CreatePlugin(ctx context.Context, pluginDto types.PluginCreateDto) (types.PluginDto, error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		type,
		title,
		description,
		metadata,
		server_endpoint,
		pricing_id
	) VALUES (
		@Type,
		@Title,
		@Description,
		@Metadata,
		@ServerEndpoint,
		@PricingID
	) RETURNING id;`, PLUGINS_TABLE)
	args := pgx.NamedArgs{
		"Type":           pluginDto.Type,
		"Title":          pluginDto.Title,
		"Description":    pluginDto.Description,
		"Metadata":       pluginDto.Metadata,
		"ServerEndpoint": pluginDto.ServerEndpoint,
		"PricingID":      pluginDto.PricingID,
	}

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return types.PluginDto{}, err
	}

	ratingQuery := `INSERT INTO plugin_rating (plugin_id, rating, count) 
          VALUES ($1, 1, 0), ($1, 2, 0), ($1, 3, 0), ($1, 4, 0), ($1, 5, 0)`

	_, err = p.pool.Exec(ctx, ratingQuery, createdId)
	if err != nil {
		return types.PluginDto{}, err
	}

	return p.FindPluginById(ctx, createdId)
}

func (p *PostgresBackend) UpdatePlugin(ctx context.Context, id string, updates types.PluginUpdateDto) (types.PluginDto, error) {
	t := reflect.TypeOf(updates)
	v := reflect.ValueOf(updates)
	numFields := t.NumField()

	query := fmt.Sprintf(`UPDATE %s SET `, PLUGINS_TABLE)
	args := pgx.NamedArgs{
		"id": id,
	}

	// iterate over dto props and assign non-empty for update
	var updateStatements []string
	for i := 0; i < numFields; i++ {
		field := t.Field(i) // field metadata
		value := v.Field(i) // field value

		// filter out json undefined values
		if !value.IsNil() {
			// get db field name (same as it is defined in the json input)
			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				// fallback to prop name
				fieldName = field.Name
			}

			// get value from dto reference
			var fieldValue interface{}
			if field.Type == reflect.TypeOf((*json.RawMessage)(nil)) {
				// keep as reference to []byte
				fieldValue = value.Interface().(*json.RawMessage)
			} else {
				// dereference
				fieldValue = value.Elem().Interface()
			}

			updateStatements = append(updateStatements, fmt.Sprintf("%s = @%s", fieldName, fieldName))
			args[fieldName] = fieldValue
		}
	}

	if len(updateStatements) == 0 {
		return types.PluginDto{}, errors.New("no updates provided")
	}

	query += strings.Join(updateStatements, ", ")
	query += " WHERE id = @id;"

	_, err := p.pool.Exec(ctx, query, args)
	if err != nil {
		return types.PluginDto{}, err
	}

	return p.FindPluginById(ctx, id)
}

func (p *PostgresBackend) DeletePluginById(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, PLUGINS_TABLE)

	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
