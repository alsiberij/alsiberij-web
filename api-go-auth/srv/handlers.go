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
)

type (
	Handler fasthttp.RequestHandler
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

	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := repository.AuthPostgresRepository.UserRepository(conn)

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

	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	refTokenRep := repository.AuthPostgresRepository.RefreshTokenRepository(conn)

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
	ctx.SetContentType("application/json")
}

func Users(ctx *fasthttp.RequestCtx) {
	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := repository.AuthPostgresRepository.UserRepository(conn)

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

	conn, err := repository.AuthPostgresRepository.AcquireConnection()
	if err != nil {
		Set500(ctx, err)
		return
	}
	defer conn.Release()

	userRep := repository.AuthPostgresRepository.UserRepository(conn)

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
		refTokenRep := repository.AuthPostgresRepository.RefreshTokenRepository(conn)

		err = refTokenRep.SetExpiredByUserId(user.Id)
		if err != nil {
			Set400(ctx, InvalidRequestBodyUserMessage)
			return
		}
	}
}
