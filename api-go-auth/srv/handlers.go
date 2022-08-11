package srv

import (
	"auth/jwt"
	"auth/repository"
	"auth/utils"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

const (
	RefreshTokenLength     = uint(1024)
	RefreshTokenAlphabet   = "-="
	RefreshTokenLifePeriod = 7 * 24 * time.Hour

	RefreshTokenRevokeTypeCurrent          = "CURRENT"
	RefreshTokenRevokeTypeAll              = "ALL"
	RefreshTokenRevokeTypeAllExceptCurrent = "ALL_EXCEPT_CURRENT"
)

type (
	Handler fasthttp.RequestHandler
)

var (
	PostgresAuth repository.Postgres
)

func Test(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(struct {
		Status bool `json:"status"`
	}{Status: true})
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
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	user, exists, err := userRep.ByCredentials(request.Login, request.Password)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if !exists {
		Set401(ctx)
		return
	}

	if user.IsBanned {
		Set400(ctx, AccountIsBannedUserMessage)
		return
	}

	refTokenRep := PostgresAuth.RefreshTokens(conn)

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
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	refTokenRep := PostgresAuth.RefreshTokens(conn)

	refreshToken, exists, err := refTokenRep.ByTokenNotExpired(request.RefreshToken)
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
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	refTokenRep := PostgresAuth.RefreshTokens(conn)

	_, exists, err := refTokenRep.ByToken(request.RefreshToken)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if !exists {
		Set400(ctx, InvalidRefreshTokenUserMessage)
		return
	}

	revokeType := string(ctx.QueryArgs().Peek("type"))

	switch revokeType {
	case RefreshTokenRevokeTypeCurrent:
		err = refTokenRep.SetExpiredByToken(request.RefreshToken)
	case RefreshTokenRevokeTypeAll:
		err = refTokenRep.SetExpiredByTokenBelongingUser(request.RefreshToken)
	case RefreshTokenRevokeTypeAllExceptCurrent:
		err = refTokenRep.SetExpiredByTokenBelongingUserExceptCurrent(request.RefreshToken)
	default:
		Set400(ctx, InvalidRevokingRefreshTokenType)
		return
	}
	if err != nil {
		Set500(ctx, err)
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
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

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

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
	ctx.SetContentType("application/json")
}

func Users(ctx *fasthttp.RequestCtx) {
	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	list, err := userRep.AllShort()
	if err != nil {
		Set500(ctx, err)
		return
	}

	_ = json.NewEncoder(ctx).Encode(UsersResponse{
		Count: len(list),
		List:  list,
	})
	ctx.SetContentType("application/json")
}

func ChangeUserStatus(ctx *fasthttp.RequestCtx) {
	userIdFromRequest := ctx.UserValue("id").(string)
	userId, err := strconv.ParseInt(userIdFromRequest, 10, 64)
	if err != nil {
		Set400(ctx, InvalidUserIdUserMessage)
		return
	}

	var request ChangeUserStatusRequest
	err = json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	conn, err := PostgresAuth.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := PostgresAuth.Users(conn)

	user, exists, err := userRep.ById(userId)
	if err != nil {
		Set500(ctx, err)
		return
	}
	if !exists {
		Set400(ctx, InvalidUserIdUserMessage)
	}

	switch ctx.UserValue(JwtContext).(jwt.Claims).Rol {
	case jwt.RoleModerator:
		if !utils.ExistsIn(jwt.CanBeBannedByModerator, user.Role) {
			Set403(ctx)
			return
		}
	case jwt.RoleAdmin:
		if !utils.ExistsIn(jwt.CanBeBannedByAdmin, user.Role) {
			Set403(ctx)
			return
		}
	case jwt.RoleCreator:
		if !utils.ExistsIn(jwt.CanBeBannedByCreator, user.Role) {
			Set403(ctx)
			return
		}
	default:
		Set403(ctx)
		return
	}

	err = userRep.ChangeStatus(user.Id, request.IsBanned)
	if err != nil {
		Set400(ctx, InvalidRequestBodyUserMessage)
		return
	}

	if request.IsBanned {
		refTokenRep := PostgresAuth.RefreshTokens(conn)

		err = refTokenRep.SetExpiredByUserId(user.Id)
		if err != nil {
			Set400(ctx, InvalidRequestBodyUserMessage)
			return
		}
	}
}
