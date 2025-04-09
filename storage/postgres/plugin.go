package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/types"
)

const REVIEWS_TABLE = "reviews"
const PLUGINS_TABLE = "plugins"

func (p *PostgresBackend) FindPluginById(ctx context.Context, dbTx pgx.Tx, id string) (*types.PluginDto, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, PLUGINS_TABLE)

	var pluginDto types.PluginDto

	err := dbTx.QueryRow(ctx, query, id).Scan(
		&pluginDto.ID,
		&pluginDto.CreatedAt,
		&pluginDto.UpdatedAt,
		&pluginDto.Type,
		&pluginDto.Title,
		&pluginDto.Description,
		&pluginDto.Metadata,
		&pluginDto.ServerEndpoint,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &pluginDto, nil
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

func (p *PostgresBackend) CreatePlugin(ctx context.Context, dbTx pgx.Tx, pluginDto types.PluginCreateDto) (string, error) {
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
	err := dbTx.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return "", fmt.Errorf("failed to insert plugin: %w", err)
	}

	return createdId, nil
}

func (p *PostgresBackend) UpdatePlugin(ctx context.Context, id string, updates types.PluginUpdateDto) (*types.PluginDto, error) {
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
		return nil, errors.New("no updates provided")
	}

	query += strings.Join(updateStatements, ", ")
	query += " WHERE id = @id;"

	_, err := p.pool.Exec(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return p.FindPluginById(ctx, nil, id)
}

func (p *PostgresBackend) DeletePluginById(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, PLUGINS_TABLE)

	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresBackend) FindReviewById(ctx context.Context, db pgx.Tx, id string) (*types.ReviewDto, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, REVIEWS_TABLE)

	var reviewDto types.ReviewDto
	err := db.QueryRow(ctx, query, id).Scan(
		&reviewDto.ID,
		&reviewDto.Address,
		&reviewDto.Rating,
		&reviewDto.Comment,
		&reviewDto.CreatedAt,
		&reviewDto.PluginId,
	)
	if err != nil {
		return nil, err
	}

	return &reviewDto, nil
}

func (p *PostgresBackend) FindReviews(ctx context.Context, pluginId string, skip int, take int, sort string) (types.ReviewsDto, error) {
	if p.pool == nil {
		return types.ReviewsDto{}, fmt.Errorf("database pool is nil")
	}

	allowedSortingColumns := map[string]bool{"created_at": true}
	orderBy, orderDirection := common.GetSortingCondition(sort, allowedSortingColumns)

	query := fmt.Sprintf(`
		SELECT *, COUNT(*) OVER() AS total_count
		FROM %s
		WHERE plugin_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, REVIEWS_TABLE, orderBy, orderDirection)

	rows, err := p.pool.Query(ctx, query, pluginId, take, skip)
	if err != nil {
		return types.ReviewsDto{}, err
	}

	defer rows.Close()

	var reviews []types.Review
	var totalCount int

	for rows.Next() {
		var review types.Review

		err := rows.Scan(
			&review.ID,
			&review.Address,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.PluginId,
			&totalCount,
		)
		if err != nil {
			return types.ReviewsDto{}, err
		}

		reviews = append(reviews, review)
	}

	pluginsDto := types.ReviewsDto{
		Reviews:    reviews,
		TotalCount: totalCount,
	}

	return pluginsDto, nil
}

func (p *PostgresBackend) CreateReview(ctx context.Context, reviewDto types.ReviewCreateDto, pluginId string) (string, error) {
	columns := []string{"address", "rating", "comment", "plugin_id", "created_at"}
	argNames := []string{"@Address", "@Rating", "@Comment", "@PluginId", "@CreatedAt"}
	args := pgx.NamedArgs{
		"Address":   reviewDto.Address,
		"Rating":    reviewDto.Rating,
		"Comment":   reviewDto.Comment,
		"PluginId":  pluginId,
		"CreatedAt": time.Now(),
	}

	query := fmt.Sprintf(
		`INSERT INTO reviews (%s) VALUES (%s) RETURNING id;`,
		strings.Join(columns, ", "),
		strings.Join(argNames, ", "),
	)

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return "", err
	}

	return createdId, nil
}
