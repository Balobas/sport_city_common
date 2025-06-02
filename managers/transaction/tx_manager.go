package transaction

import (
	"context"

	common "github.com/balobas/sport_city_common"
	clientDB "github.com/balobas/sport_city_common/clients/database"
	"github.com/balobas/sport_city_common/logger"
	"github.com/pkg/errors"
)

type Manager struct {
	dbc clientDB.ClientDB
}

func NewTxManager(client clientDB.ClientDB) *Manager {
	return &Manager{
		dbc: client,
	}
}

type Tx struct {
	transactors    []Transactor
	isolationLevel string
}

func (m *Manager) ExecuteTx(ctx context.Context, isolationLevel string, f func(ctx context.Context) error) (err error) {
	log := logger.From(ctx)
	log.Debug().Msg("txManager: execute tx call")

	if m.dbc.DB().HasTxInCtx(ctx) {
		log.Debug().Msg("txManager: tx already in context, execute in having tx")
		return errors.WithStack(f(ctx))
	}

	ctxTx, tx, err := m.dbc.DB().BeginTxWithContext(ctx, isolationLevel)
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

// Deprecated
func (m *Manager) NewPgTransaction(isolationLevel string) Tx {
	return m.NewTransaction(isolationLevel, m.dbc.DB())
}

// Deprecated
func (m *Manager) NewTransaction(isolationLevel string, transactors ...Transactor) Tx {
	return Tx{
		transactors:    transactors,
		isolationLevel: isolationLevel,
	}
}

// Deprecated
func (tx Tx) Execute(ctx context.Context, f func(ctx context.Context) error) (err error) {
	log := logger.From(ctx)
	log.Debug().Msg("txManager: execute tx call")
	internalTransactions := make([]common.Transaction, 0, len(tx.transactors))

	ctxTx := ctx
	// TODO: перенести дефер сюда, чтобы в случае нескольких реп, это соответствовало реализации 2PC
	for _, tr := range tx.transactors {
		var internalTx common.Transaction

		ctxTx, internalTx, err = tr.BeginTxWithContext(ctx, tx.isolationLevel)
		if err != nil {
			return errors.WithStack(errors.Wrap(err, "failed to begin internal tx"))
		}

		internalTransactions = append(internalTransactions, internalTx)
	}

	defer func() {
		defer func() {
			tx.transactors = nil
		}()

		if r := recover(); r != nil {
			err = errors.Wrapf(err, "panic recovered(while execute tx): %v", r)
			log.Warn().Err(err).Send()
		}

		if err != nil {
			for idx, txn := range internalTransactions {
				if !tx.transactors[idx].HasTxInCtx(ctx) {

					if rollbackErr := txn.Rollback(ctxTx); rollbackErr != nil {
						err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
					}

					log.Debug().Msg("rollback tx")
				}
			}
			return
		}

		for idx, txn := range internalTransactions {
			if !tx.transactors[idx].HasTxInCtx(ctx) {
				if commitErr := txn.Commit(ctxTx); commitErr != nil {
					err = errors.Wrapf(err, "commit error: %v", commitErr)
				}
				log.Debug().Msg("commit tx")
			}
		}
	}()

	if err := f(ctxTx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
