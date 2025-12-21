package pgRw

type pgClientOptions struct {
	readOnly bool
	maxConns int32
	minConns int32
}

type PgClientOption func(p *pgClientOptions)

func withReadOnly() func(*pgClientOptions) {
	return func(opts *pgClientOptions) {
		opts.readOnly = true
	}
}

func WithMaxConns(maxConns int32) func(*pgClientOptions) {
	return func(opts *pgClientOptions) {
		opts.maxConns = maxConns
	}
}

func WithMinConns(minConns int32) func(*pgClientOptions) {
	return func(opts *pgClientOptions) {
		opts.minConns = minConns
	}
}
