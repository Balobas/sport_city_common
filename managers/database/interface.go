package dbManager

import (
	"context"

	common "github.com/balobas/sport_city_common"
)

type ClientDB interface {
	HasTxInCtx(ctx context.Context) bool
	BeginTxWithContext(ctx context.Context, isolationLevel string) (context.Context, common.Transaction, error)
	CtxWithMasterKey(ctx context.Context) context.Context
	CtxWithReplicaKey(ctx context.Context) context.Context
}
