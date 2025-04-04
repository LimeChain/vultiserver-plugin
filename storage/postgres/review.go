package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/types"
)

const REVIEWS_TABLE = "reviews"
const PLUGIN_RATING_TABLE = "plugin_rating"

func (p *PostgresBackend) FindReviewById(ctx context.Context, id string) (types.ReviewDto, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, REVIEWS_TABLE)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return types.ReviewDto{}, err
	}

	review, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.Review])
	if err != nil {
		return types.ReviewDto{}, err
	}

	ratings, err := p.FindRatingByPluginId(ctx, review.PluginId)
	if err != nil {
		return types.ReviewDto{}, err
	}

	var reviewDto types.ReviewDto

	reviewDto.ID = review.ID
	reviewDto.Address = review.Address
	reviewDto.Comment = review.Comment
	reviewDto.CreatedAt = review.CreatedAt
	reviewDto.Rating = review.Rating
	reviewDto.PluginId = review.PluginId
	reviewDto.Ratings = ratings
	return reviewDto, nil
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

func (p *PostgresBackend) CreateReview(ctx context.Context, reviewgDto types.ReviewCreateDto, pluginId string) (types.ReviewDto, error) {
	columns := []string{"address", "rating", "comment", "plugin_id", "created_at"}
	argNames := []string{"@Address", "@Rating", "@Comment", "@PluginId", "@CreatedAt"}
	args := pgx.NamedArgs{
		"Address":   reviewgDto.Address,
		"Rating":    reviewgDto.Rating,
		"Comment":   reviewgDto.Comment,
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
		return types.ReviewDto{}, err
	}

	ratingQuery := `
	UPDATE plugin_rating
    SET count = count + 1
    WHERE plugin_id = $1 AND rating = $2`

	_, err = p.pool.Exec(ctx, ratingQuery, pluginId, reviewgDto.Rating)
	if err != nil {
		return types.ReviewDto{}, err
	}

	return p.FindReviewById(ctx, createdId)
}

func (p *PostgresBackend) FindRatingByPluginId(ctx context.Context, pluginId string) ([]types.PluginRatingDto, error) {
	query := `
	SELECT *
    FROM plugin_rating
    WHERE plugin_id = $1`

	rows, err := p.pool.Query(ctx, query, pluginId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []types.PluginRatingDto
	for rows.Next() {
		var review types.PluginRating
		err := rows.Scan(
			&review.PluginID,
			&review.Rating,
			&review.Count,
		)
		if err != nil {
			return []types.PluginRatingDto{}, err
		}

		var reviewDto types.PluginRatingDto
		reviewDto.Rating = review.Rating
		reviewDto.Count = review.Count
		ratings = append(ratings, reviewDto)
	}

	return ratings, nil
}
