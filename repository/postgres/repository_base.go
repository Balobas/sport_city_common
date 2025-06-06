package repositoryBasePostgres

import (
	"context"

	common "github.com/balobas/sport_city_common"
	clientDB "github.com/balobas/sport_city_common/clients/database"
	"github.com/balobas/sport_city_common/logger"
	"github.com/pkg/errors"
)

type BasePgRepository struct {
	dbc clientDB.ClientDB
}

func New(client clientDB.ClientDB) *BasePgRepository {
	return &BasePgRepository{
		dbc: client,
	}
}

func (r *BasePgRepository) DB() clientDB.DB {
	return r.dbc.DB()
}

func (r *BasePgRepository) BeginTxWithContext(ctx context.Context, isolationLevel string) (context.Context, common.Transaction, error) {
	return r.dbc.DB().BeginTxWithContext(ctx, isolationLevel)
}

// Deprecated
func HandleTxEnd(ctx context.Context, tx common.Transaction, err error) error {
	log := logger.From(ctx)
	if err == nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			log.Debug().Msg("with tx: failed to commit tx")
			return errors.Wrap(commitErr, "failed to commit transaction")
		}
		log.Debug().Msg("with tx: successfully commit tx")
		return nil
	}

	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		log.Debug().Msg("with tx: failed to rollback tx")
		return errors.Wrap(rollbackErr, "failed to rollback transaction")
	}
	log.Debug().Msg("with tx: successfully rollback tx")
	return nil
}
