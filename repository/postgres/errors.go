package repositoryBasePostgres

import (
	commonErrors "github.com/balobas/sport_city_common/errors"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
)

func convertError(err error) error {
	pgErr := &pgconn.PgError{}
	if errors.As(err, pgErr) {
		switch pgErr.Code {
		case "23505":
			return commonErrors.ErrAlreadyExists

		default:
		}

	}
	return err
}
