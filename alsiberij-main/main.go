package main

import (
	"app/server"
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
		port = "1409"
	}

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

	log.Fatal(s.ListenAndServe(":" + port))
}
