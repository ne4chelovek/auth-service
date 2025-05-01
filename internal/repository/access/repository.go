package access

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ne4chelovek/auth-service/internal/repository"
)

const (
	roleEndpointTable = "role_endpoints"
	roleColumn        = "role"
	endpointColumn    = "endpoint"
)

type repo struct {
	db repository.QueryRunner
}

func NewAccessRepository(db *pgxpool.Pool) repository.AccessRepository {
	return &repo{db: db}
}

func (r *repo) GetRoleEndpoints(ctx context.Context, endpoint string) ([]string, error) {
	builderSelect := sq.Select(roleColumn).
		PlaceholderFormat(sq.Dollar).
		From(roleEndpointTable).
		Where(sq.Eq{endpointColumn: endpoint})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role endpoints: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("failed to fetch role endpoints: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}
