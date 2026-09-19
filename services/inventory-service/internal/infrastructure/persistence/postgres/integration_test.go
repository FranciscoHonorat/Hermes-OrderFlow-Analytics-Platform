//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/stock"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/services/inventory-service/internal/infrastructure/persistence/postgres"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// setupDB sobe um Postgres real em container (testcontainers), aplica as
// migrations do inventory-service via golang-migrate (ADR-001) e devolve
// uma conexão pronta para uso. O container é derrubado automaticamente ao
// fim do teste que o chamou.
func setupDB(t *testing.T) *postgres.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("inventory_test"),
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

func newTestSKU(t *testing.T, value string) valueobject.SKU {
	t.Helper()
	sku, err := valueobject.NewSKU(value)
	require.NoError(t, err)
	return sku
}

// TestStockRepository exercises the Postgres adapter against a real
// database: the happy path (save, find, update via upsert) and the bad
// path (SKU that was never saved).
func TestStockRepository(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	repo := postgres.NewStockRepository(db, postgres.NewStockMapper())

	t.Run("Save and FindBySKU", func(t *testing.T) {
		sku := newTestSKU(t, "sku-integration-1")
		item, err := stock.NewStockItem(sku, 10, 2, time.Now())
		require.NoError(t, err)

		require.NoError(t, repo.Save(ctx, item))

		found, err := repo.FindBySKU(ctx, sku)
		require.NoError(t, err)
		require.Equal(t, sku.String(), found.SKU().String())
		require.Equal(t, 10, found.Available())
		require.Equal(t, 0, found.Reserved())
		require.Equal(t, 2, found.MinimumQuantity())

		// Reserve and save again: exercises the ON CONFLICT DO UPDATE branch.
		require.NoError(t, found.Reserve(4, time.Now()))
		require.NoError(t, repo.Save(ctx, found))

		updated, err := repo.FindBySKU(ctx, sku)
		require.NoError(t, err)
		require.Equal(t, 6, updated.Available())
		require.Equal(t, 4, updated.Reserved())
	})

	t.Run("FindBySKU returns error when the SKU does not exist", func(t *testing.T) {
		_, err := repo.FindBySKU(ctx, newTestSKU(t, "does-not-exist"))
		require.Error(t, err)
	})
}

// TestStockQueries exercises the read side (CQRS) against a real database.
func TestStockQueries(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	repo := postgres.NewStockRepository(db, postgres.NewStockMapper())
	queries := postgres.NewStockQueries(db)

	t.Run("GetStockBySKU and ListLowStock", func(t *testing.T) {
		healthy, err := stock.NewStockItem(newTestSKU(t, "sku-healthy"), 50, 5, time.Now())
		require.NoError(t, err)
		require.NoError(t, repo.Save(ctx, healthy))

		low, err := stock.NewStockItem(newTestSKU(t, "sku-low"), 1, 5, time.Now())
		require.NoError(t, err)
		require.NoError(t, repo.Save(ctx, low))

		dto, err := queries.GetStockBySKU(ctx, "sku-healthy")
		require.NoError(t, err)
		require.Equal(t, "sku-healthy", dto.SKU)
		require.Equal(t, 50, dto.Available)

		lowStock, err := queries.ListLowStock(ctx)
		require.NoError(t, err)
		require.Len(t, lowStock, 1)
		require.Equal(t, "sku-low", lowStock[0].SKU)
	})
}

// TestUnitOfWork exercises the Transactional Outbox path described in
// ADR-003/LEARNING-004: the happy path (aggregate and domain events
// committed atomically, then drained through shared/outbox — the piece the
// future cdc-connector will rely on) and the bad path (rollback on error).
func TestUnitOfWork(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	uow := postgres.NewUnitOfWork(db)
	repo := postgres.NewStockRepository(db, postgres.NewStockMapper())

	t.Run("commits stock and outbox atomically", func(t *testing.T) {
		sku := newTestSKU(t, "sku-uow-commit")
		item, err := stock.NewStockItem(sku, 10, 2, time.Now())
		require.NoError(t, err)
		require.NoError(t, item.Reserve(3, time.Now()))

		err = uow.Do(ctx, func(store output.RepositoryProvider) error {
			if err := store.StockRepository().Save(ctx, item); err != nil {
				return err
			}
			return store.OutboxRepository().SaveEvents(ctx, item.PullEvents())
		})
		require.NoError(t, err)

		saved, err := repo.FindBySKU(ctx, sku)
		require.NoError(t, err)
		require.Equal(t, 7, saved.Available())
		require.Equal(t, 3, saved.Reserved())

		outboxRepo := outbox.NewPostgresRepository(db.Pool)
		rows, err := outboxRepo.FetchUnprocessed(ctx, 10)
		require.NoError(t, err)

		var found *outbox.Row
		for i := range rows {
			if rows[i].AggregateID == sku.String() {
				found = &rows[i]
				break
			}
		}
		require.NotNil(t, found, "expected an outbox row for the reserved stock item")
		require.Equal(t, "stock.reserved", found.Type)

		require.NoError(t, outboxRepo.MarkProcessed(ctx, []uuid.UUID{found.ID}))

		remaining, err := outboxRepo.FetchUnprocessed(ctx, 10)
		require.NoError(t, err)
		for _, r := range remaining {
			require.NotEqual(t, found.ID, r.ID, "row should no longer be unprocessed after MarkProcessed")
		}
	})

	t.Run("rolls back on error", func(t *testing.T) {
		sku := newTestSKU(t, "sku-uow-rollback")
		item, err := stock.NewStockItem(sku, 10, 2, time.Now())
		require.NoError(t, err)

		boom := errors.New("boom")
		err = uow.Do(ctx, func(store output.RepositoryProvider) error {
			if err := store.StockRepository().Save(ctx, item); err != nil {
				return err
			}
			return boom
		})
		require.ErrorIs(t, err, boom)

		_, err = repo.FindBySKU(ctx, sku)
		require.Error(t, err, "the stock item must not have been persisted after rollback")
	})
}
