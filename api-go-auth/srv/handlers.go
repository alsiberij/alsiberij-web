package srv

import (
	"auth/database"
	"auth/jwt"
	"auth/utils"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

const (
	RefreshTokenLength         = uint(1024)
	RefreshTokenAlphabet       = `<->`
	RefreshTokenAlphabetRegexp = `^[\<\-\>]+$`
	RefreshTokenLifeTime       = 24 * time.Hour

	RefreshTokenRevokeTypeCurrent          = "CURRENT"
	RefreshTokenRevokeTypeAll              = "ALL"
	RefreshTokenRevokeTypeAllExceptCurrent = "ALL_EXCEPT_CURRENT"
)

func Test(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(TestResponse{Status: true})
	ctx.SetContentType("application/json")
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

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	userId, exists, err := userRep.IdByCredentials(request.Login, request.Password)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if !exists {
		Set401(ctx)
		return
	}

	banRep := Redis.Bans()

	ban, exists, err := banRep.ByUserId(userId)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if exists {
		userMsg := AccountIsBannedUserMessage
		userMsg.Message = fmt.Sprintf(
			userMsg.Message, ban.Reason, ban.ByUserId, ban.At, ban.Until)
		Set403WithUserMessage(ctx, userMsg)
		return
	}

	refTokenRep := PostgresAuth.RefreshTokens(conn)

	refreshToken := utils.GenerateString(RefreshTokenLength, RefreshTokenAlphabet)
	err = refTokenRep.Create(userId, refreshToken)
	if err != nil {
		Set500Error(ctx, err)
		return
	}

	_ = json.NewEncoder(ctx).Encode(LoginResponse{
		RefreshToken: refreshToken,
	})
	ctx.SetContentType("application/json")
}

func Refresh(ctx *fasthttp.RequestCtx) {
	var request RefreshRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, userMessage := request.Validate()
	if !isValid {
		Set400(ctx, userMessage)
		return
	}

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	defer conn.Release()

	refTokenRep := PostgresAuth.RefreshTokens(conn)

	tokenData, exists, err := refTokenRep.ByToken(request.RefreshToken, RefreshTokenLifeTime)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if !exists {
		Set401(ctx)
		return
	}

	token, exp, iat := jwt.Create(tokenData.UserId, tokenData.UserRole)
	response := RefreshResponse{
		JWT:       token,
		ExpiresAt: exp,
		IssuedAt:  iat,
	}

	_ = json.NewEncoder(ctx).Encode(response)
	ctx.SetContentType("application/json")
}

func Revoke(ctx *fasthttp.RequestCtx) {
	var request RefreshRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, userMessage := request.Validate()
	if !isValid {
		Set400(ctx, userMessage)
		return
	}

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	defer conn.Release()

	revokeType := string(ctx.QueryArgs().Peek("type"))

	refTokenRep := PostgresAuth.RefreshTokens(conn)

	switch revokeType {
	case RefreshTokenRevokeTypeCurrent:
		_, err = refTokenRep.RevokeToken(request.RefreshToken)
	case RefreshTokenRevokeTypeAll:
		_, err = refTokenRep.RevokeAllTokens(request.RefreshToken)
	case RefreshTokenRevokeTypeAllExceptCurrent:
		_, err = refTokenRep.RevokeAllTokensExceptOne(request.RefreshToken)
	default:
		Set400(ctx, InvalidRevokingRefreshTokenType)
		return
	}

	if err != nil {
		Set500Error(ctx, err)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
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

	//TODO REDIS
	//TODO GENERATE CODE
	code := 111111
	//TODO SEND EMAIL

	database.EmailCache.Save(request.Email, code)
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

	code, ok := database.EmailCache.Search(request.Email)
	if !(ok && code == request.Code) {
		Set400(ctx, InvalidCodeUserMessage)
		return
	}

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	exists, err := userRep.LoginExists(request.Login)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if exists {
		Set400(ctx, LoginExistsUserMessage)
		return
	}

	exists, err = userRep.EmailExists(request.Email)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if exists {
		Set400(ctx, EmailExistsUserMessage)
		return
	}

	err = userRep.Create(request.Email, request.Login, request.Password)
	if err != nil {
		Set500Error(ctx, err)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
}

func ValidateJWT(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		Set403(ctx)
		return
	}

	_ = json.NewEncoder(ctx).Encode(ValidateJwtResponse{
		JwtClaims: claims,
	})
	ctx.SetContentType("application/json")
}

func CreateBan(ctx *fasthttp.RequestCtx) {
	userIdFromRequest := ctx.UserValue("id").(string)
	userId, err := strconv.ParseInt(userIdFromRequest, 10, 64)
	if err != nil {
		Set400(ctx, InvalidUserIdUserMessage)
		return
	}

	var request BanRequest
	err = json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	isValid, userMessage := request.Validate()
	if !isValid {
		Set400(ctx, userMessage)
		return
	}

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	userRole, exists, err := userRep.RoleById(userId)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if !exists {
		Set400(ctx, InvalidUserIdUserMessage)
		return
	}

	jwtToken := ctx.UserValue(JwtContext).(jwt.Claims)

	switch jwtToken.Rol {
	case jwt.RoleModerator:
		if !utils.ExistsIn(jwt.CanBeBannedByModerator, userRole) {
			Set403(ctx)
			return
		}
	case jwt.RoleAdmin:
		if !utils.ExistsIn(jwt.CanBeBannedByAdmin, userRole) {
			Set403(ctx)
			return
		}
	case jwt.RoleCreator:
		if !utils.ExistsIn(jwt.CanBeBannedByCreator, userRole) {
			Set403(ctx)
			return
		}
	default:
		Set403(ctx)
		return
	}

	bans := Redis.Bans()

	err = bans.Create(userId, request.Reason, request.Until, jwtToken.Sub)
	if err != nil {
		Set500Error(ctx, err)
		return
	}

	refTokenRep := PostgresAuth.RefreshTokens(conn)

	_, _ = refTokenRep.RevokeAllByUserId(userId)
}

func DeleteBan(ctx *fasthttp.RequestCtx) {
	userIdFromRequest := ctx.UserValue("id").(string)
	userId, err := strconv.ParseInt(userIdFromRequest, 10, 64)
	if err != nil {
		Set400(ctx, InvalidUserIdUserMessage)
		return
	}

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	userRole, exists, err := userRep.RoleById(userId)
	if err != nil {
		Set500Error(ctx, err)
		return
	}
	if !exists {
		Set400(ctx, InvalidUserIdUserMessage)
		return
	}

	jwtToken := ctx.UserValue(JwtContext).(jwt.Claims)

	if jwtToken.Rol == jwt.RoleAdmin && userRole == jwt.RoleAdmin {
		Set403(ctx)
		return
	}

	bans := Redis.Bans()

	err = bans.Delete(userId)
	if err != nil {
		Set500Error(ctx, err)
	}
}
