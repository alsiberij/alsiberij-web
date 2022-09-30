package app

import (
	"auth/internal/models"
	"auth/internal/storages"
	"auth/pkg/jwt"
	"auth/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/valyala/fasthttp"
	"strconv"
)

func (a *Application) status(ctx *fasthttp.RequestCtx) {
	_ = json.NewEncoder(ctx).Encode(testResponse{Status: true})
	ctx.SetContentType("application/json")
}

func (a *Application) checkEmail(ctx *fasthttp.RequestCtx) {
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

	exists, err := storages.NewUserStorage(conn).EmailExists(request.Email)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if exists {
		a.setCustomError(ctx, models.EmailExistsError)
		return
	}

	code := a.rnd.Code(VerificationCodeLength)
	err = storages.NewCodeStorage(a.rdsClient1.Client()).CreateAndStore(request.Email, code, VerificationCodeLifetime)
	if err != nil {
		a.set500(ctx, err)
	}
}

func (a *Application) register(ctx *fasthttp.RequestCtx) {
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

	ok, err := storages.NewCodeStorage(a.rdsClient1.Client()).VerifyCode(request.Email, request.Code)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if !ok {
		a.setCustomError(ctx, models.WrongCodeError)
		return
	}

	conn, err := a.pgsPool.AcquireConnection(context.Background())
	if err != nil {
		a.set500(ctx, err)
		return
	}
	defer conn.Release()

	users := storages.NewUserStorage(conn)

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

func (a *Application) login(ctx *fasthttp.RequestCtx) {
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

	user, err := storages.NewUserStorage(conn).GetByCredentials(models.UserCredentials{
		Login:    request.Login,
		Password: request.Password,
	})
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.WrongCredentialsError)
		return
	}

	ban, err := storages.NewBanStorage(a.rdsClient0.Client()).Get(user.Id)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if ban != nil {
		a.set403Banned(ctx, ban)
		return
	}

	refreshToken := a.rnd.String(RefreshTokenLength, RefreshTokenAlphabet)
	err = storages.NewRefreshTokenStorage(conn).CreateAndStore(user.Id, refreshToken)
	if err != nil {
		a.set500(ctx, err)
		return
	}

	_ = json.NewEncoder(ctx).Encode(loginResponse{
		RefreshToken: refreshToken,
	})
	ctx.SetContentType("application/json")
}

func (a *Application) refresh(ctx *fasthttp.RequestCtx) {
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

	refreshToken, err := storages.NewRefreshTokenStorage(conn).Get(request.RefreshToken, RefreshTokenLifePeriod)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if refreshToken == nil {
		a.setCustomError(ctx, models.WrongRefreshTokenError)
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

func (a *Application) revoke(ctx *fasthttp.RequestCtx) {
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

	refTokens := storages.NewRefreshTokenStorage(conn)

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

func (a *Application) jwtInfo(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		a.set500(ctx, errors.New("access token error"))
		return
	}

	_ = json.NewEncoder(ctx).Encode(claims)
	ctx.SetContentType("application/json")
}

func (a *Application) ban(ctx *fasthttp.RequestCtx) {
	userIdFromRequest, _ := ctx.UserValue("id").(string)
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

	user, err := storages.NewUserStorage(conn).GetById(userId)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.WrongUserIdError)
		return
	}

	jwtToken, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		a.set500(ctx, errors.New("access token error"))
		return
	}

	myRole, ok := models.ToRole(jwtToken.Rol)
	if !ok {
		a.setCustomError(ctx, models.InvalidMyRoleError)
		return
	}

	if !myRole.IsHigher(user.Role) {
		a.setCustomError(ctx, models.NoPermissionToBanUserError)
		return
	}

	err = storages.NewBanStorage(a.rdsClient0.Client()).CreateAndStore(userId, request.Reason, request.Until, jwtToken.Sub)
	if err != nil {
		a.set500(ctx, err)
		return
	}

	_ = storages.NewRefreshTokenStorage(conn).RevokeAllByUserId(userId)
}

func (a *Application) unban(ctx *fasthttp.RequestCtx) {
	userIdFromRequest, _ := ctx.UserValue("id").(string)
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

	user, err := storages.NewUserStorage(conn).GetById(userId)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.WrongUserIdError)
		return
	}

	jwtToken, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		a.set500(ctx, errors.New("access token error"))
		return
	}

	myRole, ok := models.ToRole(jwtToken.Rol)
	if !ok {
		a.setCustomError(ctx, models.InvalidMyRoleError)
		return
	}

	if !myRole.IsHigher(user.Role) {
		a.setCustomError(ctx, models.NoPermissionToUnbanUserError)
		return
	}

	err = storages.NewBanStorage(a.rdsClient0.Client()).Delete(userId)
	if err != nil {
		a.set500(ctx, err)
	}
}

func (a *Application) changeRole(ctx *fasthttp.RequestCtx) {
	userIdFromRequest, _ := ctx.UserValue("id").(string)
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

	users := storages.NewUserStorage(conn)

	user, err := users.GetById(userId)
	if err != nil {
		a.set500(ctx, err)
		return
	}
	if user == nil {
		a.setCustomError(ctx, models.WrongUserIdError)
		return
	}

	jwtToken, ok := ctx.UserValue(JwtContext).(jwt.Claims)
	if !ok {
		a.set500(ctx, errors.New("access token error"))
		return
	}

	myRole, ok := models.ToRole(jwtToken.Rol)
	if !ok {
		a.setCustomError(ctx, models.InvalidMyRoleError)
		return
	}

	requestRole, ok := models.ToRole(request.Role)
	if !ok {
		a.setCustomError(ctx, models.InvalidRoleError)
		return
	}

	if requestRole.IsHigherOrEqual(myRole) {
		a.setCustomError(ctx, models.NoPermissionsToSetThisRoleError)
		return
	}

	if user.Role.IsHigherOrEqual(myRole) {
		a.setCustomError(ctx, models.NoPermissionToChangeUserRoleError)
		return
	}

	err = users.ChangeRole(userId, request.Role)
	if err != nil {
		a.set500(ctx, err)
	}
}
