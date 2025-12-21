package dbManager

import (
	"context"

	"github.com/balobas/sport_city_common/logger"
	"github.com/pkg/errors"
)

type Manager struct {
	dbc ClientDB
}

func NewDbManager(client ClientDB) *Manager {
	return &Manager{
		dbc: client,
	}
}

func (m *Manager) MasterCtx(ctx context.Context) context.Context {
	return m.dbc.CtxWithMasterKey(ctx)
}

func (m *Manager) ReplicaCtx(ctx context.Context) context.Context {
	return m.dbc.CtxWithReplicaKey(ctx)
}

func (m *Manager) ExecuteTx(ctx context.Context, isolationLevel string, f func(ctx context.Context) error) (err error) {
	log := logger.From(ctx)
	log.Debug().Msg("txManager: execute tx call")

	if m.dbc.HasTxInCtx(ctx) {
		log.Debug().Msg("txManager: tx already in context, execute in having tx")
		return errors.WithStack(f(ctx))
	}

	ctxTx, tx, err := m.dbc.BeginTxWithContext(ctx, isolationLevel)
	if err != nil {
		return errors.WithStack(errors.Wrap(err, "failed to begin tx"))
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrapf(err, "panic recovered(while execute tx): %v", r)
			log.Warn().Err(err).Send()
		}

		if err != nil {
			if rollbackErr := tx.Rollback(ctxTx); rollbackErr != nil {
				err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
			}

			log.Debug().Msg("rollback tx")
			return
		}

		if commitErr := tx.Commit(ctxTx); commitErr != nil {
			err = errors.Wrapf(err, "commit error: %v", commitErr)
		}
		log.Debug().Msg("commit tx")
	}()

	if err := f(ctxTx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
