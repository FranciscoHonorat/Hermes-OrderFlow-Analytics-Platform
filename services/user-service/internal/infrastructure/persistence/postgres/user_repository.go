package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/input"
	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type UserRepository struct {
	q      DBTX
	mapper *UserMapper
}

func NewUserRepository(db *DB, mapper *UserMapper) repository.UserRepository {
	return &UserRepository{q: db.Pool, mapper: mapper}
}

func NewUserRepositoryFromTx(tx pgx.Tx, mapper *UserMapper) repository.UserRepository {
	return &UserRepository{q: tx, mapper: mapper}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	row := r.mapper.ToPersistence(u)

	query := `
		INSERT INTO users (id, email, password_hash, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at;
	`

	_, err := r.q.Exec(ctx, query,
		row.ID,
		row.Email,
		row.PasswordHash,
		row.Role,
		row.Active,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return domainErrors.ErrEmailAlreadyRegistered
		}
		return fmt.Errorf("failed to save user to postgres: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id valueobject.UserID) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, role, active, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var row UserRow

	err := r.q.QueryRow(ctx, query, id.String()).Scan(
		&row.ID,
		&row.Email,
		&row.PasswordHash,
		&row.Role,
		&row.Active,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user %s: %w", id.String(), err)
	}

	return r.mapper.ToDomain(&row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email valueobject.Email) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, role, active, created_at, updated_at
		FROM users
		WHERE email = $1;
	`

	var row UserRow

	err := r.q.QueryRow(ctx, query, email.String()).Scan(
		&row.ID,
		&row.Email,
		&row.PasswordHash,
		&row.Role,
		&row.Active,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email %s: %w", email.String(), err)
	}

	return r.mapper.ToDomain(&row)
}

type UserQueries struct {
	pool DBTX
}

var _ input.UserQueries = (*UserQueries)(nil)

func NewUserQueries(db *DB) *UserQueries {
	return &UserQueries{pool: db.Pool}
}

func (q *UserQueries) GetUserByID(ctx context.Context, id string) (*input.UserDTO, error) {
	query := `
		SELECT id, email, role, active, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var dto input.UserDTO
	if err := q.pool.QueryRow(ctx, query, id).Scan(
		&dto.ID,
		&dto.Email,
		&dto.Role,
		&dto.Active,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("query get user by id failed: %w", err)
	}

	return &dto, nil
}

func (q *UserQueries) ListUsers(ctx context.Context) ([]input.UserDTO, error) {
	query := `
		SELECT id, email, role, active, created_at, updated_at
		FROM users
		ORDER BY created_at;
	`

	dbRows, err := q.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query list users failed: %w", err)
	}
	defer dbRows.Close()

	rows := make([]input.UserDTO, 0)
	for dbRows.Next() {
		var dto input.UserDTO
		if err := dbRows.Scan(&dto.ID, &dto.Email, &dto.Role, &dto.Active, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		rows = append(rows, dto)
	}

	return rows, nil
}
