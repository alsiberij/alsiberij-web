package app

import (
	"auth/internal/models"
	"auth/internal/storage"
	"auth/pkg/jwt"
	"auth/pkg/utils"
	"context"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"strconv"
)

func (a *application) status(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(testResponse{Status: true})
	ctx.SetContentType("application/json")
}

func (a *application) checkEmail(ctx *fasthttp.RequestCtx) {
	var request checkEmailRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storage.NewUserStorage(conn)

	exists, err := users.EmailExists(request.Email)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if exists {
		a.setCustomError(ctx, models.EmailExistsError)
	}

	codes := storage.NewCodeStorage(a.rdsClient1.Client())

	code := a.rnd.Code(VerificationCodeLength)
	err = codes.CreateAndStore(request.Email, code, VerificationCodeLifetime)
	if err != nil {
		a.set500(ctx, err)
	}
}

func (a *application) register(ctx *fasthttp.RequestCtx) {
	var request registerRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	codes := storage.NewCodeStorage(a.rdsClient1.Client())

	ok, err := codes.Verify(request.Email, request.Code)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if !ok {
		a.setCustomError(ctx, models.InvalidCodeError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storage.NewUserStorage(conn)

	exists, err := users.LoginExists(request.Login)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if exists {
		a.setCustomError(ctx, models.LoginExistsError)
		return
	}

	err = users.CreateAndStore(request.Email, request.Login, request.Password)
	if err != nil {
		a.set500(ctx, err)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
}

func (a *application) login(ctx *fasthttp.RequestCtx) {
	var request loginRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storage.NewUserStorage(conn)

	user, err := users.GetByCredentials(models.UserCredentials{Login: request.Login, Password: request.Password})
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.InvalidLoginOrPasswordError)
		return
	}

	bans := storage.NewBanStorage(a.rdsClient0.Client())

	ban, err := bans.Get(user.Id)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if ban != nil {
		a.setCustomError(ctx, models.AccountIsBannedError)
		return
	}

	refTokens := storage.NewRefreshTokenStorage(conn)

	refreshToken := a.rnd.String(RefreshTokenLength, RefreshTokenAlphabet)
	err = refTokens.CreateAndStore(user.Id, refreshToken)
	if err != nil {
		a.set500(ctx, err)
		return
	}

	_ = json.NewEncoder(ctx).Encode(loginResponse{
		RefreshToken: refreshToken,
	})
	ctx.SetContentType("application/json")
}

func (a *application) refresh(ctx *fasthttp.RequestCtx) {
	var request refreshRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	refTokens := storage.NewRefreshTokenStorage(conn)

	refreshToken, err := refTokens.Get(request.RefreshToken, RefreshTokenLifePeriod)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if refreshToken == nil {
		a.setCustomError(ctx, models.InvalidRefreshTokenError)
		return
	}

	accessToken, exp, iat := jwt.Create(refreshToken.User.Id, string(refreshToken.User.Role))
	response := refreshResponse{
		AccessToken: accessToken,
		ExpiresAt:   exp,
		IssuedAt:    iat,
	}

	_ = json.NewEncoder(ctx).Encode(response)
	ctx.SetContentType("application/json")
}

func (a *application) revoke(ctx *fasthttp.RequestCtx) {
	var request refreshRequest
	err := json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	revokeType := string(ctx.QueryArgs().Peek("type"))

	if !utils.ExistsIn(revokeTypes, revokeType) {
		a.setCustomError(ctx, models.InvalidRevokeTypeError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	refTokens := storage.NewRefreshTokenStorage(conn)

	switch revokeType {
	case RefreshTokenRevokeTypeCurrent:
		err = refTokens.Revoke(request.RefreshToken)
	case RefreshTokenRevokeTypeAll:
		err = refTokens.RevokeAll(request.RefreshToken)
	case RefreshTokenRevokeTypeAllExceptCurrent:
		err = refTokens.RevokeAllExceptCurrent(request.RefreshToken)
	default:
		a.setCustomError(ctx, models.InvalidRevokeTypeError)
		return
	}

	if err != nil {
		a.set500(ctx, err)
	}

}

func (a *application) jwtInfo(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		a.set403(ctx)
		return
	}

	_ = json.NewEncoder(ctx).Encode(claims)
	ctx.SetContentType("application/json")
}

func (a *application) ban(ctx *fasthttp.RequestCtx) {
	userIdFromRequest := ctx.UserValue("id").(string)
	userId, err := strconv.ParseInt(userIdFromRequest, 10, 64)
	if err != nil {
		a.set400(ctx)
		return
	}

	var request banRequest
	err = json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storage.NewUserStorage(conn)

	user, err := users.GetById(userId)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.InvalidUserIdError)
		return
	}

	jwtToken, _ := ctx.UserValue(JwtContext).(jwt.Claims)
	myRole, ok := models.ToRole(jwtToken.Rol)
	if !ok {
		a.set403(ctx)
		return
	}

	if !myRole.IsHigher(user.Role) {
		a.set403(ctx)
		return
	}

	bans := storage.NewBanStorage(a.rdsClient0.Client())

	err = bans.CreateAndStore(userId, request.Reason, request.Until, jwtToken.Sub)
	if err != nil {
		a.set500(ctx, err)
		return
	}

	refTokens := storage.NewRefreshTokenStorage(conn)

	_ = refTokens.RevokeAllByUserId(userId)
}

func (a *application) unban(ctx *fasthttp.RequestCtx) {
	userIdFromRequest := ctx.UserValue("id").(string)
	userId, err := strconv.ParseInt(userIdFromRequest, 10, 64)
	if err != nil {
		a.set400(ctx)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storage.NewUserStorage(conn)

	user, err := users.GetById(userId)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.InvalidUserIdError)
		return
	}

	jwtToken, _ := ctx.UserValue(JwtContext).(jwt.Claims)
	myRole, ok := models.ToRole(jwtToken.Rol)
	if !ok {
		a.set403(ctx)
		return
	}

	if !myRole.IsHigher(user.Role) {
		a.setCustomError(ctx, models.InvalidUserIdNoPermission)
		return
	}

	bans := storage.NewBanStorage(a.rdsClient0.Client())

	err = bans.Delete(userId)
	if err != nil {
		a.set500(ctx, err)
	}
}

func (a *application) changeRole(ctx *fasthttp.RequestCtx) {
	userIdFromRequest := ctx.UserValue("id").(string)
	userId, err := strconv.ParseInt(userIdFromRequest, 10, 64)
	if err != nil {
		a.set400(ctx)
		return
	}

	var request changeRoleRequest
	err = json.Unmarshal(ctx.Request.Body(), &request)
	if err != nil {
		a.set400(ctx)
		return
	}

	requestError, err := request.Validate()
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if requestError != nil {
		a.setCustomError(ctx, requestError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storage.NewUserStorage(conn)

	user, err := users.GetById(userId)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.InvalidUserIdError)
		return
	}

	jwtToken, _ := ctx.UserValue(JwtContext).(jwt.Claims)
	myRole, _ := models.ToRole(jwtToken.Rol)
	requestRole, _ := models.ToRole(request.Role)

	if requestRole.IsHigher(myRole) || user.Role.IsHigherOrEqual(myRole) {
		a.set403(ctx)
		return
	}

	err = users.ChangeRole(userId, request.Role)
	if err != nil {
		a.set500(ctx, err)
	}
}
