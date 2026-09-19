package postgres

import (
	"context"

	sharedpg "github.com/FranciscoHonorat/ordemflow/shared/postgres"
)

type DB = sharedpg.DB

func NewConnection(ctx context.Context, dsn string) (*DB, error) {
	return sharedpg.NewConnection(ctx, dsn)
}
