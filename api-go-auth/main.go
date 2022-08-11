package main

import (
	"auth/jwt"
	"auth/logger"
	"auth/repository"
	"auth/srv"
	"crypto/tls"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"os"
	"time"
)

const (
	V1 = "/v1"
)

//TODO graceful shutdown, redis cache, line 32 refactor

func init() {
	config, err := ReadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	pgs, err := repository.New(config.AuthPGS)
	if err != nil {
		log.Fatalf("UNABLE CONNECT TO POSTGRES: %s", err.Error())
	}
	srv.PostgresAuth = pgs

	logsPath := os.Getenv("LOGS_PATH")
	if logsPath == "" {
		logsPath = "./logger/logs"
	}
	logger.LogsPath = logsPath

	go repository.EmailCache.GC()
}

func main() {
	r := router.New()
	r.RedirectTrailingSlash = false

	r.NotFound = srv.Set404
	r.MethodNotAllowed = srv.Set405
	r.PanicHandler = srv.Set500Panic

	r.GET(V1+"/", srv.Test)

	r.POST(V1+"/login", srv.Login)

	r.POST(V1+"/refresh", srv.Refresh)

	r.DELETE(V1+"/refresh", srv.Revoke)

	r.POST(V1+"/checkEmail", srv.CheckEmail)

	r.POST(V1+"/register", srv.Register)

	r.GET(V1+"/validateJWT", srv.WithMiddlewares(srv.ValidateJWT, srv.Authorize))

	r.GET(V1+"/users", srv.WithMiddlewares(srv.Users,
		srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator})))

	r.PATCH(V1+"/user/{id}/status", srv.WithMiddlewares(srv.ChangeUserStatus,
		srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator})))

	portSec := os.Getenv("PORT")
	if portSec == "" {
		portSec = "11400"
	}
	log.Printf("LISTENING SECURE %s PORT\n", portSec)

	sslPath := os.Getenv("SSL_PATH")
	if sslPath == "" {
		sslPath = "./ssl"
	}

	cert, err := tls.LoadX509KeyPair(sslPath+"/fullchain.pem", sslPath+"/privkey.pem")
	if err != nil {
		log.Fatalf("SSL ERROR: %s", err.Error())
	}

	lisSec, err := tls.Listen("tcp4", ":"+portSec, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		log.Fatalf("SECURE LISTENER ERROR: %s", err.Error())
	}

	serverSecure := fasthttp.Server{
		Name:         "API-GO-AUTH",
		Handler:      srv.LogMiddleware(r.Handler),
		LogAllErrors: true,
	}

	errorsStream := make(chan error)

	go Serve(&serverSecure, lisSec, errorsStream)

	for {
		log.Println(<-errorsStream)
	}
}

func Serve(server *fasthttp.Server, listener net.Listener, errChan chan error) {
	for {
		err := server.Serve(listener)
		errChan <- err
		time.Sleep(5 * time.Second)
	}
}
