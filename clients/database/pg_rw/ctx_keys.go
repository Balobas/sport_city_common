package pgRw

import "context"

type (
	TxKey        struct{}
	PgMasterKey  struct{}
	PgReplicaKey struct{}
)

func (p *pgClientRw) CtxWithMasterKey(ctx context.Context) context.Context {
	return context.WithValue(ctx, PgMasterKey{}, PgMasterKey{})
}

func (p *pgClientRw) CtxWithReplicaKey(ctx context.Context) context.Context {
	return context.WithValue(ctx, PgReplicaKey{}, PgReplicaKey{})
}

func (p *pgClientRw) getConnByCtxKey(ctx context.Context) PgConn {
	if val := ctx.Value(PgMasterKey{}); val != nil {
		return p.masterPool
	}

	if val := ctx.Value(PgReplicaKey{}); val != nil {
		return p.replicaPool
	}

	return p.replicaPool
}
