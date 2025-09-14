package shutdown

import (
	"context"
	"log"
	"sync"
)

var cl closer

type closer struct {
	mx    sync.Mutex
	once  sync.Once
	funcs []CloserFunc
}

type CloserFunc func(ctx context.Context) error

func Add(f ...CloserFunc) {
	cl.mx.Lock()
	cl.funcs = append(cl.funcs, f...)
	cl.mx.Unlock()
}

func CloseAll(ctx context.Context) {
	cl.mx.Lock()
	defer cl.mx.Unlock()

	cl.once.Do(func() {
		errs := make(chan error, len(cl.funcs))

		go func() {
			for i := len(cl.funcs) - 1; i >= 0; i-- {
				errs <- cl.funcs[i](ctx)
			}
			close(errs)
		}()

		done := make(chan struct{}, 1)
		defer close(done)

		go func(errs chan error, done chan struct{}) {
			for err := range errs {
				if err != nil {
					log.Printf("failed to close: %v", err)
				}
			}
			done <- struct{}{}
		}(errs, done)

		select {
		case <-done:
			log.Printf("shutdown successfuly finished")
		case <-ctx.Done():
			select {
			case <-done:
				log.Printf("shutdown successfuly finished")
				return
			default:
			}
			log.Printf("failed to finish shutdown: %v", ctx.Err())
			return
		}
	})
}

func WrapClose(f func(ctx context.Context)) CloserFunc {
	return func(ctx context.Context) error {
		f(ctx)
		return nil
	}
}
