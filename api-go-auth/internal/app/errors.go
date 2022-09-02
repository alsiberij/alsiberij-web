package app

import (
	"auth/internal/models"
	"auth/pkg/logging"
	"auth/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
)

type (
	appError struct {
		StatusCode   int    `json:"statusCode"`
		DevMsg       string `json:"devMsg"`
		UsrMsg       string `json:"usrMsg"`
		InternalCode int    `json:"internalCode"`
	}
)

func (a *application) setCustomError(ctx *fasthttp.RequestCtx, err *models.Error) {
	e := convertError(err)
	_ = json.NewEncoder(ctx).Encode(e)
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(e.StatusCode)
}

func (a *application) set400(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode: fasthttp.StatusBadRequest,
		DevMsg:     "Bad request",
		UsrMsg:     "Bad request",
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
}

func (a *application) set401(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode: fasthttp.StatusUnauthorized,
		DevMsg:     "Unauthorized",
		UsrMsg:     "Unauthorized",
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
}

func (a *application) set403(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode:   fasthttp.StatusForbidden,
		DevMsg:       "Forbidden",
		UsrMsg:       "Forbidden",
		InternalCode: 1,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusForbidden)
}

func (a *application) set403Banned(ctx *fasthttp.RequestCtx, ban *models.Ban) {
	var usrMsg string
	if ban != nil {
		usrMsg = fmt.Sprintf("Your account was banned (%s - %s) by user #%d by reason: %s",
			ban.At.Format("15:04 02-01-2006"),
			ban.Until.Format("15:04 02-01-2006"),
			ban.ByUserId,
			ban.Reason)
	}
	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode:   fasthttp.StatusForbidden,
		DevMsg:       "Account is banned",
		UsrMsg:       usrMsg,
		InternalCode: 1,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusForbidden)
}

func (a *application) set404(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode:   fasthttp.StatusNotFound,
		DevMsg:       "Not found: " + utils.BytesToString(ctx.Path()),
		UsrMsg:       "Not found",
		InternalCode: 1,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusNotFound)
}

func (a *application) set405(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode:   fasthttp.StatusMethodNotAllowed,
		DevMsg:       "Method not allowed",
		UsrMsg:       "Method not allowed",
		InternalCode: 1,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
}

func (a *application) set500(ctx *fasthttp.RequestCtx, err error) {
	var devMsg string
	if err != nil {
		go a.logger.LogError(err, logging.LevelError)
		devMsg += err.Error()
	} else {
		devMsg += "Empty error"
	}

	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode:   fasthttp.StatusInternalServerError,
		DevMsg:       devMsg,
		UsrMsg:       "Internal server error",
		InternalCode: 1,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
}

func (a *application) set500Fatal(ctx *fasthttp.RequestCtx, i interface{}) {
	var devMsg string

	switch T := i.(type) {
	case error:
		devMsg += T.Error()
	case string:
		devMsg += T
	case nil:
		devMsg += "Empty fatal error"
	default:
		devMsg += "Unknown fatal error"
	}

	go a.logger.LogError(errors.New(devMsg), logging.LevelFatal)

	_ = json.NewEncoder(ctx).Encode(appError{
		StatusCode:   fasthttp.StatusInternalServerError,
		DevMsg:       "Internal server fatal error",
		UsrMsg:       "Internal server error",
		InternalCode: 1,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
}
