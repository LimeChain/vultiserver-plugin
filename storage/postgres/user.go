package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/internal/types"
)

const USERS_TABLE = "users"

func (p *PostgresBackend) FindUserByCredentials(
	ctx context.Context,
	username string,
	passwordHash string,
) (*types.User, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE username = $1 AND password = $2 LIMIT 1;`, USERS_TABLE)

	args := []interface{}{username, passwordHash}
	rows, err := p.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}
