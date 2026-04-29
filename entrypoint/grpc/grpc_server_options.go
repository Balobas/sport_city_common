package entrypointGrpc

import (
	loggingInterceptor "github.com/balobas/sport_city_common/grpc/interceptors/logging"
	"google.golang.org/grpc"
)

type ServerOptions struct {
	withoutReflection bool

	withAuthInterceptor    bool
	authInterceptorOptions AuthInterceptorOptions

	withLoggingInterceptor bool
	loggingInterceptorOpts []loggingInterceptor.Option

	withTracingInterceptor bool

	withCustomInterceptors bool
	customInterceptors     []grpc.UnaryServerInterceptor
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{}
}

func (so *ServerOptions) Apply(opts ...ServerOption) {
	for _, apply := range opts {
		apply(so)
	}
}

func (so *ServerOptions) WithoutReflection() bool {
	return so.withoutReflection
}

func (so *ServerOptions) WithAuthInterceptor() bool {
	return so.withAuthInterceptor
}

func (so *ServerOptions) AuthInterceptorOptions() AuthInterceptorOptions {
	return so.authInterceptorOptions
}

func (so *ServerOptions) WithLoggingInterceptor() bool {
	return so.withLoggingInterceptor
}

func (so *ServerOptions) WithTracingInterceptor() bool {
	return so.withTracingInterceptor
}

func (so *ServerOptions) LoggingInterceptorOptions() []loggingInterceptor.Option {
	return so.loggingInterceptorOpts
}

func (so *ServerOptions) WithCustomInterceptors() bool {
	return so.withCustomInterceptors
}

func (so *ServerOptions) CustomInterceptors() []grpc.UnaryServerInterceptor {
	return so.customInterceptors
}

type AuthInterceptorOptions struct {
	withoutAuthMap map[string]struct{}
}

func (aio *AuthInterceptorOptions) WithoutAuthMap() map[string]struct{} {
	return aio.withoutAuthMap
}

type ServerOption func(*ServerOptions)

func WithoutReflection() ServerOption {
	return func(so *ServerOptions) {
		so.withoutReflection = true
	}
}

func WithAuthInterceptor(withoutAuthMap map[string]struct{}) ServerOption {
	return func(so *ServerOptions) {
		so.withAuthInterceptor = true
		if withoutAuthMap == nil {
			so.authInterceptorOptions.withoutAuthMap = map[string]struct{}{}
		} else {
			so.authInterceptorOptions.withoutAuthMap = withoutAuthMap
		}
	}
}

func WithLoggingInterceptor(opts ...loggingInterceptor.Option) ServerOption {
	return func(so *ServerOptions) {
		so.withLoggingInterceptor = true
		so.loggingInterceptorOpts = opts
	}
}

func WithTracingInterceptor() ServerOption {
	return func(so *ServerOptions) {
		so.withTracingInterceptor = true
	}
}

func WithCustomInterceptors(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(so *ServerOptions) {
		so.withCustomInterceptors = true
		so.customInterceptors = interceptors
	}
}
