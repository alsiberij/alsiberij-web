package srv

import (
	"auth/jwt"
	"fmt"
	"github.com/valyala/fasthttp"
	"time"
)

const (
	JwtContext = "JWT_CONTEXT"
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

func AddExecutionTimeHeader(h Handler) Handler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		h(ctx)
		ctx.Response.Header.Add("Execution-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	}
}

func Authorize(h Handler) Handler {
	return func(ctx *fasthttp.RequestCtx) {
		authorization := string(ctx.Request.Header.Peek("Authorization"))
		if len(authorization) < 7 {
			Set401(ctx)
			return
		}

		bearerToken := authorization[7:]
		_, claims, err := jwt.Parse(bearerToken)
		if err != nil {
			Set401(ctx)
			return
		}

		ctx.SetUserValue(JwtContext, claims)
		h(ctx)
	}
}
