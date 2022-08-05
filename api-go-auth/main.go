package main

import (
	"auth/jwt"
	"auth/repository"
	"auth/srv"
	"crypto/tls"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
	"os"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "11400"
	}
	log.Printf("LISTENING %s PORT\n", port)

	sslPath := os.Getenv("SSL_PATH")
	if sslPath == "" {
		sslPath = "./ssl"
	}

	cert, err := tls.LoadX509KeyPair(sslPath+"/fullchain.pem", sslPath+"/privkey.pem")
	if err != nil {
		log.Fatalf("SSL ERROR: %s\n", err.Error())
	}

	lis, err := tls.Listen("tcp4", ":"+port, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		log.Fatalf("SSL ERROR: %s\n", err.Error())
	}

	r := router.New()

	r.NotFound = srv.Set404
	r.MethodNotAllowed = srv.Set405
	r.PanicHandler = srv.Set500

	r.GET("/", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Test, srv.AddExecutionTimeHeader)))

	r.POST("/login", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Login, srv.AddExecutionTimeHeader)))

	r.POST("/refresh", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Refresh, srv.AddExecutionTimeHeader)))

	r.DELETE("/refresh", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Revoke, srv.AddExecutionTimeHeader)))

	r.POST("/checkEmail", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.CheckEmail, srv.AddExecutionTimeHeader)))

	r.POST("/register", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Register, srv.AddExecutionTimeHeader)))

	r.GET("/validateJWT", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.ValidateJWT, srv.Authorize, srv.AddExecutionTimeHeader)))

	r.GET("/users", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.Users, srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator}), srv.AddExecutionTimeHeader)))

	r.PATCH("/user/{id}/status", fasthttp.RequestHandler(
		srv.WithMiddlewares(srv.ChangeUserStatus, srv.AuthorizeRoles([]string{jwt.RoleCreator, jwt.RoleAdmin, jwt.RoleModerator}), srv.AddExecutionTimeHeader)))

	s := fasthttp.Server{
		Name:    "ALSIBERIJ-API-AUTH",
		Handler: r.Handler,
	}

	log.Fatal(s.Serve(lis))
}
