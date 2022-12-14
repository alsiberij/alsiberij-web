package app

import (
	"auth/internal/models"
	"auth/internal/storages"
	"auth/pkg/jwt"
	"auth/pkg/logging"
	"auth/pkg/utils"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
	"time"
)

const (
	JwtContext = "JWT_CONTEXT"
)

func (a *Application) logMiddleware(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
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

		err := a.logger.WriteServerRequest(req, res)
		if err != nil {
			log.Printf("LOG ERROR: %v\n", err)
		}
	}
}

func (a *Application) authorize(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
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

		ban, err := storages.NewBanStorage(a.rdsClient0.Client()).Get(claims.Sub)
		if err != nil {
			a.set500(ctx, err)
			return
		}
		if ban != nil {
			a.set403Banned(ctx, ban)
			return
		}

		ctx.SetUserValue(JwtContext, claims)
		handler(ctx)
	}
}

func (a *Application) authorizeRoles(roles ...models.UserRole) middleware {
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
				a.setCustomError(ctx, models.InvalidMyRoleError)
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

			ban, err := storages.NewBanStorage(a.rdsClient0.Client()).Get(claims.Sub)
			if err != nil {
				a.set500(ctx, err)
				return
			}
			if ban != nil {
				a.set403Banned(ctx, ban)
				return
			}

			ctx.SetUserValue(JwtContext, claims)
			handler(ctx)
		}
	}
}
