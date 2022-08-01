package srv

import (
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
)

func Set400(ctx *fasthttp.RequestCtx, userMessage UserMessage) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusBadRequest,
		DevMsg:   "",
		UsrMsg:   userMessage,
	})
	ctx.SetContentType("application/json")
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
}

func Set404(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(HttpError{
		HttpCode: fasthttp.StatusNotFound,
		DevMsg:   "Not found",
		UsrMsg: UserMessage{
			Message:      "Не найдено",
			InternalCode: 1,
		},
	})
	ctx.SetContentType("application/json")
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
		UsrMsg: UserMessage{
			Message:      "Внутренняя ошибка сервера",
			InternalCode: 2,
		},
	})
	ctx.SetContentType("application/json")
}
