package srv

import (
	"auth/logger"
	"encoding/json"
	"github.com/valyala/fasthttp"
)

type (
	HttpError struct {
		HttpCode int         `json:"statusCode"`
		DevMsg   string      `json:"devMsg"`
		UsrMsg   UserMessage `json:"usrMsg"`
	}

	UserMessage struct {
		Message      string `json:"message"`
		InternalCode int    `json:"internalCode"`
	}
)

var (
	InvalidRequestBodyUserMessage = UserMessage{
		Message:      "Неверное тело запроса",
		InternalCode: -1,
	}
	InvalidLoginUserMessage = UserMessage{
		Message:      "Неверный логин",
		InternalCode: -2,
	}
	InvalidPasswordUserMessage = UserMessage{
		Message:      "Неверный пароль",
		InternalCode: -3,
	}
	InvalidEmailUserMessage = UserMessage{
		Message:      "Неправильный email",
		InternalCode: -4,
	}
	InvalidRefreshTokenUserMessage = UserMessage{
		Message:      "Неверный токен обновления",
		InternalCode: -5,
	}
	InvalidCodeUserMessage = UserMessage{
		Message:      "Неверный код из письма",
		InternalCode: -6,
	}
	LoginExistsUserMessage = UserMessage{
		Message:      "Пользователь с таким логином уже существует",
		InternalCode: -7,
	}
	EmailExistsUserMessage = UserMessage{
		Message:      "Пользователь с такой почтой уже существует",
		InternalCode: -8,
	}
	AccountIsBannedUserMessage = UserMessage{
		Message:      "Ваш аккаунт заблокирован",
		InternalCode: -9,
	}
	InvalidUserIdUserMessage = UserMessage{
		Message:      "Неверный идентификатор пользователя",
		InternalCode: -10,
	}
	InvalidRevokingRefreshTokenType = UserMessage{
		Message:      "Неверный тип отзыва токена обновления",
		InternalCode: -11,
	}
)

func Set400(ctx *fasthttp.RequestCtx, userMessage UserMessage) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusBadRequest,
		DevMsg:   "",
		UsrMsg:   userMessage,
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
}

func Set401(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusUnauthorized,
		DevMsg:   "Unauthorized",
		UsrMsg: UserMessage{
			Message:      "Не авторизован",
			InternalCode: 1,
		},
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
}

func Set403(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusForbidden,
		DevMsg:   "Forbidden",
		UsrMsg: UserMessage{
			Message:      "Доступ запрещен",
			InternalCode: 1,
		},
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusForbidden)
}

func Set404(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusNotFound,
		DevMsg:   "Not found",
		UsrMsg: UserMessage{
			Message:      "Не найдено : " + string(ctx.Path()),
			InternalCode: 1,
		},
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusNotFound)
}

func Set405(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusMethodNotAllowed,
		DevMsg:   "Method not allowed",
		UsrMsg: UserMessage{
			Message:      "Метод не поддерживается",
			InternalCode: 1,
		},
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
}

func Set500Error(ctx *fasthttp.RequestCtx, err error) {
	devMsg := "ERROR : "
	if err != nil {
		devMsg += err.Error()
		go logger.LogError(err, logger.LevelError)
	} else {
		devMsg += "empty error"
	}

	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusInternalServerError,
		DevMsg:   devMsg,
		UsrMsg: UserMessage{
			Message:      "Внутренняя ошибка сервера",
			InternalCode: 2,
		},
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
}

func Set500Panic(ctx *fasthttp.RequestCtx, i interface{}) {
	devMsg := "PANIC : "

	switch T := i.(type) {
	case error:
		devMsg += T.Error()
	case string:
		devMsg += T
	case nil:
		devMsg += "nil"
	default:
		devMsg += "Unknown error"
	}

	go logger.LogMessage(devMsg, logger.LevelFatal)

	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusInternalServerError,
		DevMsg:   devMsg,
		UsrMsg: UserMessage{
			Message:      "Внутренняя ошибка сервера",
			InternalCode: 2,
		},
	})
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
}
