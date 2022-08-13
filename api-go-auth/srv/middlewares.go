package srv

import (
	"auth/jwt"
	"auth/logging"
	"auth/utils"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

const (
	JwtContext = "JWT_CONTEXT"
)

type (
	Middleware func(handler fasthttp.RequestHandler) fasthttp.RequestHandler
)

func WithMiddlewares(h fasthttp.RequestHandler, mds ...Middleware) fasthttp.RequestHandler {
	handler := h
	for i := range mds {
		handler = mds[i](handler)
	}
	return handler
}

func Authorize(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		authorization := string(ctx.Request.Header.Peek("Authorization"))
		_, bearerToken, ok := strings.Cut(authorization, "Bearer ")
		if !ok {
			Set401(ctx)
			return
		}

		_, claims, err := jwt.Parse(bearerToken)
		if err != nil {
			Set401(ctx)
			return
		}

		ctx.SetUserValue(JwtContext, claims)
		h(ctx)
	}
}

func AuthorizeRoles(roles []string) Middleware {
	return func(h fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			authorization := string(ctx.Request.Header.Peek("Authorization"))
			_, bearerToken, ok := strings.Cut(authorization, "Bearer ")
			if !ok {
				Set401(ctx)
				return
			}

			_, claims, err := jwt.Parse(bearerToken)
			if err != nil {
				Set401(ctx)
				return
			}

			if !utils.ExistsIn(roles, claims.Rol) {
				Set403(ctx)
				return
			}

			ctx.SetUserValue(JwtContext, claims)
			h(ctx)
		}
	}
}

func LogMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		req := logging.Request{
			Timestamp: time.Now().Unix(),
			Method:    utils.BytesToString(ctx.Request.Header.Method()),
			Path:      utils.BytesToString(ctx.Path()),
			Protocol:  utils.BytesToString(ctx.Request.Header.Protocol()),
			Headers:   make([]string, 0, ctx.Request.Header.Len()),
			Body:      utils.BytesToString(ctx.Request.Body()),
		}

		ctx.Request.Header.VisitAll(func(key, value []byte) {
			req.Headers = append(req.Headers,
				utils.BytesToString(append(append(key, []byte{':', ' '}...), value...)))
		})

		t1 := time.Now()
		h(ctx)
		t2 := time.Now()

		res := logging.Response{
			Timestamp:     time.Now().Unix(),
			Protocol:      utils.BytesToString(ctx.Response.Header.Protocol()),
			StatusCode:    ctx.Response.StatusCode(),
			Headers:       make([]string, 0, ctx.Response.Header.Len()),
			Body:          utils.BytesToString(ctx.Response.Body()),
			ExecutionTime: t2.Sub(t1).Milliseconds(),
		}
		ctx.Response.Header.VisitAll(func(key, value []byte) {
			res.Headers = append(res.Headers,
				utils.BytesToString(append(append(key, []byte{':', ' '}...), value...)))
		})

		go Logger.WriteServerRequest(req, res)
	}
}
