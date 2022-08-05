package main

import (
	"auth/jwt"
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
	ApiV1 = "/api/v1"
)

func init() {
	config, err := ReadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = repository.AuthPostgresRepository.Init(config.AuthPG)
	if err != nil {
		log.Fatal(err)
	}

	go repository.EmailCache.GC()
}

func main() {
	r := router.New()
	r.RedirectTrailingSlash = false

	r.NotFound = srv.Set404
	r.MethodNotAllowed = srv.Set405
	r.PanicHandler = srv.Set500

	r.GET(ApiV1+"/", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Test, srv.AddExecutionTimeHeader)))

	r.POST(ApiV1+"/login", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Login, srv.AddExecutionTimeHeader)))

	r.POST(ApiV1+"/refresh", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Refresh, srv.AddExecutionTimeHeader)))

	r.DELETE(ApiV1+"/refresh", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Revoke, srv.AddExecutionTimeHeader)))

	r.POST(ApiV1+"/checkEmail", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.CheckEmail, srv.AddExecutionTimeHeader)))

	r.POST(ApiV1+"/register", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Register, srv.AddExecutionTimeHeader)))

	r.GET(ApiV1+"/validateJWT", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.ValidateJWT, srv.Authorize, srv.AddExecutionTimeHeader)))

	r.GET(ApiV1+"/users", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Users, srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator}), srv.AddExecutionTimeHeader)))

	r.PATCH(ApiV1+"/user/{id}/status", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.ChangeUserStatus, srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator}), srv.AddExecutionTimeHeader)))

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
		Name:    "API-GO-AUTH:SECURE",
		Handler: r.Handler,
	}

	go Serve(&serverSecure, lisSec, errorsStream)

	lisInsec, err := net.Listen("tcp4", ":"+portInsec)
	if err != nil {
		log.Fatalf("INSECURE LISTENER ERROR: %s", err.Error())
	}

	serverInsecure := fasthttp.Server{
		Name:    "API-GO-AUTH:INSECURE",
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
