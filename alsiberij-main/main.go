package main

import (
	"app/server"
	"crypto/tls"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
	"os"
)

func init() {
	server.PathToView = os.Getenv("PATH_VIEW")
	if server.PathToView == "" {
		server.PathToView = "./view"
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "11400"
	}

	sslPath := os.Getenv("SSL_PATH")
	if sslPath == "" {
		sslPath = "./ssl"
	}

	cert, err := tls.LoadX509KeyPair(sslPath+"/fullchain.pem", sslPath+"/privkey.pem")
	if err != nil {
		log.Fatalf("NO SSL CERTIFICATES: %s", err.Error())
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", ":"+port, config)
	if err != nil {
		log.Fatalf("FAILED TO LISTEN :%s - %s", port, err.Error())
	}
	defer ln.Close()

	r := router.New()

	r.NotFound = server.Set404HTML
	r.MethodNotAllowed = server.Set405HTML
	r.PanicHandler = server.Set500HTML

	r.GET("/", server.MainHandler)

	r.GET("/favicon.ico", server.IconHandler)
	r.GET("/css/{cssFile:^[\\w-]+\\.css$}", server.CssHandler)
	r.GET("/img/{imgFile:^[\\w-]+\\.png$}", server.ImageHandler)

	s := fasthttp.Server{
		Handler: r.Handler,
		Name:    "alsiberij-main",
	}

	log.Fatal(s.Serve(ln))
}
