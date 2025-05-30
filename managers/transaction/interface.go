package transaction

import (
	"context"

	common "github.com/balobas/sport_city_common"
)

type Transactor interface {
	BeginTxWithContext(ctx context.Context, isolationLevel string) (context.Context, common.Transaction, error)
	HasTxInCtx(ctx context.Context) bool
}
