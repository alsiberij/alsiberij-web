package srv

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

const (
	InvalidRequestBodyCode    = -1
	InvalidRequestBodyMessage = "Неверное тело запроса"

	InvalidEmailCodeCode    = -2
	InvalidEmailCodeMessage = "Неверный код из письма"

	LoginExistsCode    = -3
	LoginExistsMessage = "Пользователь с таким логином уже существует"

	EmailExistsCode    = -4
	EmailExistsMessage = "Пользователь с такой почтой уже существует"
)

type (
	HttpError struct {
		InternalCode int    `json:"internalCode"`
		HttpCode     int    `json:"statusCode"`
		DevMsg       string `json:"devMsg"`
		UsrMsg       string `json:"usrMsg"`
	}
)

func Set400(ctx *fasthttp.RequestCtx, userMsg string, internalCode int) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		InternalCode: internalCode,
		HttpCode:     fasthttp.StatusBadRequest,
		DevMsg:       "",
		UsrMsg:       userMsg,
	})
	ctx.SetContentType("application/json")
}

func Set401(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		InternalCode: 0,
		HttpCode:     fasthttp.StatusUnauthorized,
		DevMsg:       "Unauthorized",
		UsrMsg:       "Не авторизован",
	})
	ctx.SetContentType("application/json")
}

func Set404(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		InternalCode: 0,
		HttpCode:     fasthttp.StatusNotFound,
		DevMsg:       "Not found",
		UsrMsg:       "Не найдено",
	})
	ctx.SetContentType("application/json")
}

func Set405(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		InternalCode: 0,
		HttpCode:     fasthttp.StatusNotFound,
		DevMsg:       "Method not allowed",
		UsrMsg:       "Метод не поддерживается",
	})
	ctx.SetContentType("application/json")
}

func Set500(ctx *fasthttp.RequestCtx, i interface{}) {
	var devMsg string
	switch T := i.(type) {
	case error:
		devMsg = T.Error()
	}
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusInternalServerError,
		DevMsg:   devMsg,
		UsrMsg:   "",
	})
	ctx.SetContentType("application/json")
}
