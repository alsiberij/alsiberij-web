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
	RefreshTokenLength     = uint(1024)
	RefreshTokenAlphabet   = "-="
	RefreshTokenLifePeriod = 7 * 24 * time.Hour
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
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, _ := request.Validate()
	if !isValid {
		Set401(ctx)
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
	err = refTokenRep.Create(user.Id, refreshToken)
	if err != nil {
		Set500(ctx, err)
		return
	}

	response := LoginResponse{
		RefreshToken: refreshToken,
	}

	_ = json.NewEncoder(ctx).Encode(response)
}

func Refresh(ctx *fasthttp.RequestCtx) {
	var request RefreshRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, _ := request.Validate()
	if !isValid {
		Set401(ctx)
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

	if time.Now().Sub(refreshToken.LastUsedAt) > RefreshTokenLifePeriod {
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
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, userMsg := request.Validate()
	if !isValid {
		Set400(ctx, userMsg)
		return
	}

	//TODO GENERATE CODE
	code := 111111
	//TODO SEND EMAIL

	repository.EmailCache.Save(request.Email, code)
}

func Register(ctx *fasthttp.RequestCtx) {
	var request RegisterRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, userMsg := request.Validate()
	if !isValid {
		Set400(ctx, userMsg)
		return
	}

	code, ok := repository.EmailCache.Search(request.Email)
	if !(ok && code == request.Code) {
		Set400(ctx, InvalidCodeUserMessage)
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
		Set400(ctx, LoginExistsUserMessage)
		return
	}

	exists, err = userRep.EmailExists(request.Email)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if exists {
		Set400(ctx, EmailExistsUserMessage)
		return
	}

	err = userRep.Create(request.Email, request.Login, request.Password)
	if err != nil {
		Set500(ctx, err)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
}

func ValidateJWT(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		Set500(ctx, claims)
		return
	}

	_ = json.NewEncoder(ctx).Encode(struct {
		Status    bool       `json:"status"`
		JwtClaims jwt.Claims `json:"jwtClaims"`
	}{
		Status:    true,
		JwtClaims: claims,
	})
}
