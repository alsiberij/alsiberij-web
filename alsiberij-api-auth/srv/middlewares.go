package srv

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"time"
)

type (
	Middleware func(Handler) Handler
)

func WithMiddlewares(h Handler, mds ...Middleware) Handler {
	handler := h
	for i := range mds {
		handler = mds[i](handler)
	}
	return handler
}

func AddJsonContentTypeHeader(h Handler) Handler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		h(ctx)
	}
}

func AddExecutionTimeHeader(h Handler) Handler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		h(ctx)
		ctx.Response.Header.Add("Execution-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	}
}
