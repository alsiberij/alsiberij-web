package srv

import (
	"auth/jwt"
	"auth/repository"
	"auth/utils"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"time"
)

const (
	RefreshTokenLength   = uint(1024)
	RefreshTokenAlphabet = "oOIiLlJj01"
	RefreshTokenLifetime = 7 * 24 * time.Hour
)

type (
	Handler fasthttp.RequestHandler
)

func Test(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(struct {
		Status bool `json:"status"`
	}{Status: true})
}

func Login(ctx *fasthttp.RequestCtx) {
	var request LoginRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyMessage, InvalidRequestBodyCode)
		return
	}

	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := repository.AuthPostgresRepository.UserRepository(conn)

	user, exists, err := userRep.GetByCredentials(request.Login, request.Password)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if !exists {
		Set401(ctx)
		return
	}

	refTokenRep := repository.AuthPostgresRepository.RefreshTokenRepository(conn)

	refreshToken := utils.GenerateString(RefreshTokenLength, RefreshTokenAlphabet)
	expiration := time.Now().Add(RefreshTokenLifetime)
	err = refTokenRep.Create(user.Id, refreshToken, expiration)
	if err != nil {
		Set500(ctx, err)
		return
	}

	response := LoginResponse{
		RefreshToken: refreshToken,
		ExpiresIn:    expiration.Unix(),
	}

	_ = json.NewEncoder(ctx).Encode(response)
}

func Refresh(ctx *fasthttp.RequestCtx) {
	var request RefreshRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyMessage, InvalidRequestBodyCode)
		return
	}

	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	refTokenRep := repository.AuthPostgresRepository.RefreshTokenRepository(conn)

	refreshToken, exists, err := refTokenRep.ByToken(request.RefreshToken)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if !exists {
		Set401(ctx)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		err = refTokenRep.SetExpired(refreshToken.Id)
		if err != nil {
			Set500(ctx, err)
		} else {
			Set401(ctx)
		}
		return
	}

	err = refTokenRep.UpdateLastUsageTime(refreshToken.Token)
	if err != nil {
		Set500(ctx, err)
		return
	}

	token, exp, iat := jwt.Create(refreshToken.User.Id, refreshToken.User.Role)
	response := RefreshResponse{
		JWT:       token,
		ExpiresAt: exp,
		IssuedAt:  iat,
	}

	_ = json.NewEncoder(ctx).Encode(response)
}

func CheckEmail(ctx *fasthttp.RequestCtx) {
	var request CheckEmailRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyMessage, InvalidRequestBodyCode)
		return
	}

	//TODO SEND EMAIL
	//TODO GENERATE CODE
	code := 111111

	repository.EmailCache.Save(request.Email, code)
}

func Register(ctx *fasthttp.RequestCtx) {
	var request RegisterRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyMessage, InvalidRequestBodyCode)
		return
	}

	code, ok := repository.EmailCache.Search(request.Email)
	if !(ok && code == request.Code) {
		Set400(ctx, InvalidEmailCodeMessage, InvalidEmailCodeCode)
		return
	}

	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := repository.AuthPostgresRepository.UserRepository(conn)

	exists, err := userRep.LoginExists(request.Login)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if exists {
		Set400(ctx, LoginExistsMessage, LoginExistsCode)
		return
	}

	exists, err = userRep.EmailExists(request.Email)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if exists {
		Set400(ctx, EmailExistsMessage, EmailExistsCode)
		return
	}

	err = userRep.Create(request.Email, request.Login, request.Password)
	if err != nil {
		Set500(ctx, err)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
}
