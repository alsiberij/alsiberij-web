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

func init() {
	config, err := ReadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = repository.AuthPostgresRepository.Init(config.AuthPG)
	if err != nil {
		log.Fatal("UNABLE CONNECT TO POSTGRES")
	}

	err = logger.Init(config.Elasticsearch)
	if err != nil {
		log.Fatal("UNABLE TO CONNECT TO ELASTICSEARCH")
	}

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
	r.PanicHandler = srv.Set500

	r.GET(V1+"/", srv.WithMiddlewares(srv.Test, srv.LogMiddleware))

	r.POST(V1+"/login", srv.WithMiddlewares(srv.Login, srv.LogMiddleware))

	r.POST(V1+"/refresh", srv.WithMiddlewares(srv.Refresh, srv.LogMiddleware))

	r.DELETE(V1+"/refresh", srv.WithMiddlewares(srv.Revoke, srv.LogMiddleware))

	r.POST(V1+"/checkEmail", srv.WithMiddlewares(srv.CheckEmail, srv.LogMiddleware))

	r.POST(V1+"/register", srv.WithMiddlewares(srv.Register, srv.LogMiddleware))

	r.GET(V1+"/validateJWT", srv.WithMiddlewares(srv.ValidateJWT, srv.Authorize, srv.LogMiddleware))

	r.GET(V1+"/users", srv.WithMiddlewares(srv.Users,
		srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator}), srv.LogMiddleware))

	r.PATCH(V1+"/user/{id}/status", srv.WithMiddlewares(srv.ChangeUserStatus,
		srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator}), srv.LogMiddleware))

	errorsStream := make(chan error)

	portSec := os.Getenv("PORT")
	if portSec == "" {
		portSec = "11400"
	}
	log.Printf("LISTENING SECURE %s PORT\n", portSec)

	portInsec := os.Getenv("PORT_INSEC")
	if portInsec == "" {
		portInsec = "10400"
	}
	log.Printf("LISTENING INSECURE %s PORT\n", portInsec)

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
		Name:    "API-GO-AUTH-SECURE",
		Handler: r.Handler,
	}

	go Serve(&serverSecure, lisSec, errorsStream)

	lisInsec, err := net.Listen("tcp4", ":"+portInsec)
	if err != nil {
		log.Fatalf("INSECURE LISTENER ERROR: %s", err.Error())
	}

	serverInsecure := fasthttp.Server{
		Name:    "API-GO-AUTH-INSECURE",
		Handler: r.Handler,
	}

	go Serve(&serverInsecure, lisInsec, errorsStream)

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
