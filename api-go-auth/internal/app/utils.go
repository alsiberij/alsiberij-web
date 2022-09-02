package app

import (
	"auth/internal/models"
	"github.com/valyala/fasthttp"
)

type (
	middleware func(handler fasthttp.RequestHandler) fasthttp.RequestHandler
)

func convertError(serviceError *models.Error) appError {
	if serviceError == nil {
		return appError{
			StatusCode:   fasthttp.StatusInternalServerError,
			DevMsg:       "Unknown error",
			UsrMsg:       "Unknown error",
			InternalCode: 1,
		}
	}

	e := appError{
		DevMsg:       serviceError.Message,
		UsrMsg:       serviceError.Message,
		InternalCode: serviceError.InnerCode,
	}

	var statusCode int
	switch serviceError.InnerCode {
	case models.WrongCredentials, models.WrongRefreshToken:

		statusCode = fasthttp.StatusUnauthorized

	case models.AccountIsBanned, models.InvalidMyRole,
		models.NoPermissionToBanUser, models.NoPermissionToUnbanUser,
		models.NoPermissionsToSetThisRole, models.NoPermissionToChangeUserRole:

		statusCode = fasthttp.StatusForbidden

	case models.WrongUserId:

		statusCode = fasthttp.StatusNotFound

	default:
		statusCode = fasthttp.StatusBadRequest
	}

	e.StatusCode = statusCode

	return e
}

func withMiddlewares(h fasthttp.RequestHandler, mds ...middleware) fasthttp.RequestHandler {
	handler := h
	for i := range mds {
		handler = mds[i](handler)
	}
	return handler
}
