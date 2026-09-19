package postgres

import (
	"context"

	sharedpg "github.com/FranciscoHonorat/ordemflow/shared/postgres"
)

// DB e NewConnection são um re-export do shared kernel (ver ADR-005): a
// conexão Postgres é idêntica em todo bounded context, então a
// implementação vive uma única vez em shared/postgres.
type DB = sharedpg.DB

func NewConnection(ctx context.Context, dsn string) (*DB, error) {
	return sharedpg.NewConnection(ctx, dsn)
}
