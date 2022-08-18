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

		bans := Redis0.Bans()

		_, exists, err := bans.Get(claims.Sub)
		if err != nil {
			Set500Error(ctx, err)
			return
		}
		if exists {
			Set403WithUserMessage(ctx, AccountIsBannedUserMessage)
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

			bans := Redis0.Bans()

			_, exists, err := bans.Get(claims.Sub)
			if err != nil {
				Set500Error(ctx, err)
				return
			}
			if exists {
				Set403WithUserMessage(ctx, AccountIsBannedUserMessage)
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
			Body:      utils.BytesToString(ctx.Request.Body()),
		}
		req.Headers = strings.Split(strings.ReplaceAll(ctx.Request.Header.String(), "\r", ""), "\n")[1:]

		t1 := time.Now()
		h(ctx)
		t2 := time.Now()

		res := logging.Response{
			Timestamp:     time.Now().Unix(),
			Protocol:      utils.BytesToString(ctx.Response.Header.Protocol()),
			StatusCode:    ctx.Response.StatusCode(),
			Body:          utils.BytesToString(ctx.Response.Body()),
			ExecutionTime: t2.Sub(t1).Milliseconds(),
		}
		res.Headers = strings.Split(strings.ReplaceAll(ctx.Response.Header.String(), "\r", ""), "\n")[1:]

		go Logger.WriteServerRequest(req, res)
	}
}
