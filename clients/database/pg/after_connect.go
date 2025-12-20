package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func setReadOnly(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, "set session characteristics as transaction read only")
	if err != nil {
		return errors.Errorf("failed to setReadonly on pg session: %v", err)
	}
	return nil
}
