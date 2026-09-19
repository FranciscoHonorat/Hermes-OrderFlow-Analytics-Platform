//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/persistence/postgres"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *postgres.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("users_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	runMigrations(t, dsn)

	db, err := postgres.NewConnection(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	return db
}

func runMigrations(t *testing.T, dsn string) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsPath := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "migrations")

	m, err := migrate.New("file://"+migrationsPath, dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		srcErr, dbErr := m.Close()
		require.NoError(t, srcErr)
		require.NoError(t, dbErr)
	})

	require.NoError(t, m.Up())
}

func newTestUser(t *testing.T, email string) *user.User {
	t.Helper()
	id := valueobject.NewUserIDMust(uuid.New())
	e, err := valueobject.NewEmail(email)
	require.NoError(t, err)

	u, err := user.NewUser(id, e, "hashed-password", valueobject.NewRoleMust("user"), time.Now())
	require.NoError(t, err)
	return u
}

func TestUserRepository(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	repo := postgres.NewUserRepository(db, postgres.NewUserMapper())

	t.Run("Save and FindByID and FindByEmail", func(t *testing.T) {
		u := newTestUser(t, "integration-1@example.com")

		require.NoError(t, repo.Save(ctx, u))

		byID, err := repo.FindByID(ctx, u.ID())
		require.NoError(t, err)
		require.Equal(t, u.Email().String(), byID.Email().String())

		byEmail, err := repo.FindByEmail(ctx, u.Email())
		require.NoError(t, err)
		require.Equal(t, u.ID().String(), byEmail.ID().String())

		require.NoError(t, byID.ChangeRole(valueobject.NewRoleMust("admin"), time.Now()))
		require.NoError(t, repo.Save(ctx, byID))

		updated, err := repo.FindByID(ctx, u.ID())
		require.NoError(t, err)
		require.Equal(t, "admin", updated.Role().String())
	})

	t.Run("Save returns ErrEmailAlreadyRegistered on duplicate email", func(t *testing.T) {
		email := "integration-dup@example.com"
		first := newTestUser(t, email)
		require.NoError(t, repo.Save(ctx, first))

		second := newTestUser(t, email)
		err := repo.Save(ctx, second)
		require.ErrorIs(t, err, domainErrors.ErrEmailAlreadyRegistered)
	})

	t.Run("FindByID returns error when the user does not exist", func(t *testing.T) {
		_, err := repo.FindByID(ctx, valueobject.NewUserIDMust(uuid.New()))
		require.Error(t, err)
	})
}

func TestUserQueries(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	repo := postgres.NewUserRepository(db, postgres.NewUserMapper())
	queries := postgres.NewUserQueries(db)

	t.Run("GetUserByID and ListUsers", func(t *testing.T) {
		u := newTestUser(t, "integration-queries@example.com")
		require.NoError(t, repo.Save(ctx, u))

		dto, err := queries.GetUserByID(ctx, u.ID().String())
		require.NoError(t, err)
		require.Equal(t, u.Email().String(), dto.Email)

		all, err := queries.ListUsers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, all)
	})
}

func TestUnitOfWork(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	uow := postgres.NewUnitOfWork(db)
	repo := postgres.NewUserRepository(db, postgres.NewUserMapper())

	t.Run("commits user and outbox atomically", func(t *testing.T) {
		u := newTestUser(t, "integration-uow-commit@example.com")

		err := uow.Do(ctx, func(store output.RepositoryProvider) error {
			if err := store.UserRepository().Save(ctx, u); err != nil {
				return err
			}
			return store.OutboxRepository().SaveEvents(ctx, u.PullEvents())
		})
		require.NoError(t, err)

		saved, err := repo.FindByID(ctx, u.ID())
		require.NoError(t, err)
		require.Equal(t, u.Email().String(), saved.Email().String())

		outboxRepo := outbox.NewPostgresRepository(db.Pool)
		rows, err := outboxRepo.FetchUnprocessed(ctx, 10)
		require.NoError(t, err)

		var found *outbox.Row
		for i := range rows {
			if rows[i].AggregateID == u.ID().String() {
				found = &rows[i]
				break
			}
		}
		require.NotNil(t, found, "expected an outbox row for the registered user")
		require.Equal(t, "user.registered", found.Type)

		require.NoError(t, outboxRepo.MarkProcessed(ctx, []uuid.UUID{found.ID}))
	})

	t.Run("rolls back on error", func(t *testing.T) {
		u := newTestUser(t, "integration-uow-rollback@example.com")

		boom := errors.New("boom")
		err := uow.Do(ctx, func(store output.RepositoryProvider) error {
			if err := store.UserRepository().Save(ctx, u); err != nil {
				return err
			}
			return boom
		})
		require.ErrorIs(t, err, boom)

		_, err = repo.FindByID(ctx, u.ID())
		require.Error(t, err, "the user must not have been persisted after rollback")
	})
}
