package app

import (
	"auth/internal/models"
	"auth/internal/storage"
	"auth/pkg/jwt"
	"auth/pkg/logging"
	"auth/pkg/utils"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

const (
	JwtContext = "JWT_CONTEXT"
)

func (a *application) logMiddleware(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		req := logging.Request{
			Timestamp: time.Now().Unix(),
			Method:    utils.BytesToString(ctx.Request.Header.Method()),
			Path:      utils.BytesToString(ctx.Path()),
			Protocol:  utils.BytesToString(ctx.Request.Header.Protocol()),
			Body:      utils.BytesToString(ctx.Request.Body()),
		}
		req.Headers = strings.Split(strings.Trim(ctx.Request.Header.String(), "\r\n"), "\r\n")[1:]

		t1 := time.Now()
		handler(ctx)
		t2 := time.Now()

		res := logging.Response{
			Timestamp:     time.Now().Unix(),
			Protocol:      utils.BytesToString(ctx.Response.Header.Protocol()),
			StatusCode:    ctx.Response.StatusCode(),
			Body:          utils.BytesToString(ctx.Response.Body()),
			ExecutionTime: t2.Sub(t1).Milliseconds(),
		}
		res.Headers = strings.Split(strings.Trim(ctx.Response.Header.String(), "\r\n"), "\r\n")[1:]

		go a.logger.WriteServerRequest(req, res)
	}
}

func (a *application) authorize(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		authorization := string(ctx.Request.Header.Peek("Authorization"))
		_, bearerToken, ok := strings.Cut(authorization, "Bearer ")
		if !ok {
			a.set401(ctx)
			return
		}

		_, claims, err := jwt.Parse(bearerToken)
		if err != nil {
			a.set401(ctx)
			return
		}

		bans := storage.NewBanStorage(a.rdsClient0.Client())

		ban, err := bans.Get(claims.Sub)
		if err != nil {
			a.set500(ctx, err)
			return
		}
		if ban != nil {
			a.setCustomError(ctx, models.AccountIsBannedError)
			return
		}

		ctx.SetUserValue(JwtContext, claims)
		handler(ctx)
	}
}

func (a *application) authorizeRoles(roles ...models.UserRole) middleware {
	return func(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			authorization := string(ctx.Request.Header.Peek("Authorization"))
			_, bearerToken, ok := strings.Cut(authorization, "Bearer ")
			if !ok {
				a.set401(ctx)
				return
			}

			_, claims, err := jwt.Parse(bearerToken)
			if err != nil {
				a.set401(ctx)
				return
			}

			myRole, ok := models.ToRole(claims.Rol)
			if !ok {
				a.set403(ctx)
				return
			}

			var found bool
			for _, role := range roles {
				if myRole == role {
					found = true
					break
				}
			}
			if !found {
				a.set403(ctx)
				return
			}

			bans := storage.NewBanStorage(a.rdsClient0.Client())

			ban, err := bans.Get(claims.Sub)
			if err != nil {
				a.set500(ctx, err)
				return
			}
			if ban != nil {
				a.setCustomError(ctx, models.AccountIsBannedError)
				return
			}

			ctx.SetUserValue(JwtContext, claims)
			handler(ctx)
		}
	}
}
