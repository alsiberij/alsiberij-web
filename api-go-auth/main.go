package main

import (
	"auth/jwt"
	"auth/logging"
	"auth/repository"
	"auth/srv"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
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

//TODO redis cache

//TODO tests

//TODO base64encodeToString, headers visiting (memory)

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
		logsPath = "./logging/logs"
	}
	srv.Logger = logging.NewLogger(fmt.Sprintf(logsPath+"/logs-%s.log", time.Now().Format("2006-01-02")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666, "2006-01-02T15:04:05")

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

	//TODO REMOVE DEBUG
	r.GET("/debug/pprof/profile", pprofhandler.PprofHandler)
	r.GET("/debug/pprof/heap", pprofhandler.PprofHandler)

	portSec := os.Getenv("PORT")
	if portSec == "" {
		portSec = "11400"
	}
	log.Printf("LISTENING %s PORT\n", portSec)

	sslPath := os.Getenv("SSL_PATH")
	if sslPath == "" {
		sslPath = "./ssl"
	}

	cert, err := tls.LoadX509KeyPair(sslPath+"/fullchain.pem", sslPath+"/privkey.pem")
	if err != nil {
		log.Fatalf("SSL ERROR: %s", err.Error())
	}

	lis, err := tls.Listen("tcp4", ":"+portSec, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		log.Fatalf("LISTENER ERROR: %s", err.Error())
	}

	server := fasthttp.Server{
		Name:         "API-GO-AUTH",
		Handler:      srv.LogMiddleware(r.Handler),
		LogAllErrors: true,
	}

	GracefulServe(&server, lis)
}

func GracefulServe(server *fasthttp.Server, listener net.Listener) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go Serve(server, listener, cancel)

	var err error

	select {
	case <-ctx.Done():
		log.Println("SERVER STOPPED")
	case <-sigChan:
		log.Println("SHUTTING DOWN...")
	}

	err = server.Shutdown()
	err = srv.Logger.Save()
	if err != nil {
		log.Printf("ERROR SAVING LOG BUFFER: %v", err)
	}

	if err != nil {
		log.Printf("SHUT DOWN ERROR: %v\n", err)
	} else {
		log.Println("SHUT DOWN OK")
	}

	return
}

func Serve(server *fasthttp.Server, listener net.Listener, cancel context.CancelFunc) {
	err := server.Serve(listener)
	if err != nil {
		log.Printf("SERVER ERROR: %v\n", err)
	}
	cancel()
}
