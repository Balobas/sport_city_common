package httpUtils

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/balobas/sport_city_common/tracer"
	"github.com/go-chi/chi"
)

type ChiTraceRouter struct {
	chi.Router
}

func ChiRouterWithTracing(r *chi.Mux) chi.Router {
	return &ChiTraceRouter{r}
}

func (c *ChiTraceRouter) HandleFunc(pattern string, h http.HandlerFunc) {
	c.Router.HandleFunc(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) MethodFunc(method, pattern string, h http.HandlerFunc) {
	c.Router.MethodFunc(method, pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Connect(pattern string, h http.HandlerFunc) {
	c.Router.Connect(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Delete(pattern string, h http.HandlerFunc) {
	c.Router.Delete(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Get(pattern string, h http.HandlerFunc) {
	c.Router.Get(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Head(pattern string, h http.HandlerFunc) {
	c.Router.Head(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Options(pattern string, h http.HandlerFunc) {
	c.Router.Options(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Patch(pattern string, h http.HandlerFunc) {
	c.Router.Patch(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Post(pattern string, h http.HandlerFunc) {
	c.Router.Post(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Put(pattern string, h http.HandlerFunc) {
	c.Router.Put(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) Trace(pattern string, h http.HandlerFunc) {
	c.Router.Trace(pattern, handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) NotFound(h http.HandlerFunc) {
	c.Router.NotFound(handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) MethodNotAllowed(h http.HandlerFunc) {
	c.Router.MethodNotAllowed(handlerFnWithTraceSpan(h))
}

func (c *ChiTraceRouter) With(middlewares ...func(http.Handler) http.Handler) chi.Router {
	r := c.Router.With(middlewares...)
	return &ChiTraceRouter{r}
}

func (c *ChiTraceRouter) Group(fn func(r chi.Router)) chi.Router {
	r := c.Router.Group(func(r chi.Router) {
		fn(&ChiTraceRouter{r})
	})
	return &ChiTraceRouter{r}
}

func (c *ChiTraceRouter) Route(pattern string, fn func(r chi.Router)) chi.Router {
	r := c.Router.Route(pattern, func(r chi.Router) {
		fn(&ChiTraceRouter{r})
	})
	return &ChiTraceRouter{r}
}

func handlerFnWithTraceSpan(fn http.HandlerFunc) http.HandlerFunc {
	var (
		parentStructName string
		fnName           string
	)
	fnOrMethodRealName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()

	parentStructParts := strings.Split(fnOrMethodRealName, "(*")
	if len(parentStructParts) != 0 {
		parentStructName = strings.Split(parentStructParts[1], ")")[0]
	}

	fnNameParts := strings.Split(fnOrMethodRealName, ".")
	if len(fnNameParts) != 0 {
		fnName = strings.Split(fnNameParts[len(fnNameParts)-1], "-")[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.FromCtx(r.Context()).Start(r.Context(), fmt.Sprintf("%s.%s", parentStructName, fnName))
		defer span.End()

		r = r.WithContext(ctx)
		fn(w, r)
	}
}
