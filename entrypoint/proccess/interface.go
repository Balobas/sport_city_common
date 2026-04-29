package entrypointProccess

import "context"

type Process interface {
	Name() string
	Run(ctx context.Context) error
}
