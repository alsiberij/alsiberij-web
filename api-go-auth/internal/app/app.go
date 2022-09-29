package app

import (
	"auth/internal/models"
	"auth/pkg/logging"
	"auth/pkg/pgs"
	"auth/pkg/rds"
	"auth/pkg/utils"
	"context"
	"errors"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	V1 = "/v1"
)

type (
	application struct {
		logger     *logging.Logger
		pgsPool    *pgs.Postgres
		rdsClient0 *rds.Redis
		rdsClient1 *rds.Redis
		server     *fasthttp.Server
		lis        net.Listener
		rnd        utils.Random
	}

	App interface {
		Serve()
	}
)

func NewApp(serverName string, logger *logging.Logger, pgsPool *pgs.Postgres, rdsClient0, rdsClient1 *rds.Redis, lis net.Listener) (App, error) {
	if logger == nil || pgsPool == nil || rdsClient0 == nil || rdsClient1 == nil || lis == nil {
		return nil, errors.New("app init error - nil arguments passed")
	}

	app := &application{
		logger:     logger,
		pgsPool:    pgsPool,
		rdsClient0: rdsClient0,
		rdsClient1: rdsClient1,
		lis:        lis,
		rnd:        utils.NewRandom(time.Now().Unix()),
	}

	r := router.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.PanicHandler = app.set500Fatal
	r.NotFound = app.set404
	r.MethodNotAllowed = app.set405

	r.GET(V1+"/", app.status)
	r.POST(V1+"/checkEmail", app.checkEmail)
	r.POST(V1+"/register", app.register)
	r.POST(V1+"/login", app.login)
	r.POST(V1+"/refresh", app.refresh)
	r.DELETE(V1+"/refresh", withMiddlewares(app.revoke, app.authorize))
	r.GET(V1+"/me/accessToken", withMiddlewares(app.jwtInfo, app.authorize))
	r.POST(V1+"/user/{id}/ban", withMiddlewares(app.ban, app.authorizeRoles(models.RoleCreator, models.RoleAdministrator)))
	r.DELETE(V1+"/user/{id}/ban", withMiddlewares(app.unban, app.authorizeRoles(models.RoleCreator, models.RoleAdministrator)))
	r.PATCH(V1+"/user/{id}/role", withMiddlewares(app.changeRole, app.authorizeRoles(models.RoleCreator, models.RoleAdministrator)))

	server := &fasthttp.Server{
		Handler: app.logMiddleware(r.Handler),
		Name:    serverName,
	}

	app.server = server

	return app, nil
}

func (a *application) Serve() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(server *fasthttp.Server, listener net.Listener, cancel context.CancelFunc) {
		err := server.Serve(listener)
		if err != nil {
			log.Printf("SERVER ERROR: %v\n", err)
		}
		cancel()
	}(a.server, a.lis, cancel)

	var err error

	select {
	case <-ctx.Done():
		log.Println("SERVER STOPPED")
	case <-sigChan:
		log.Println("SHUTTING DOWN...")
	}

	err = a.server.Shutdown()
	if err != nil {
		log.Printf("SHUT DOWN ERROR: %v\n", err)
	}

	err = a.logger.Close()
	if err != nil {
		log.Printf("ERROR SAVING LOG BUFFER: %v\n", err)
	}

	if err != nil {
		log.Printf("SHUT DOWN ERROR: %v\n", err)
	} else {
		log.Println("SHUT DOWN OK")
	}
}
