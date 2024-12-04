package common

import "context"

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

const (
	Serializable    string = "serializable"
	RepeatableRead  string = "repeatable read"
	ReadCommitted   string = "read committed"
	ReadUncommitted string = "read uncommitted"
)
