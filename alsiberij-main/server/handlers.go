package server

import (
	"github.com/valyala/fasthttp"
	"html/template"
	"io"
	"log"
	"os"
)

var (
	PathToView string
)

func MainHandler(ctx *fasthttp.RequestCtx) {
	tmpl, err := template.ParseFiles(PathToView + "/html/index.html")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}
	ctx.SetContentType(ContentTypeHtml)
	err = tmpl.Execute(ctx, nil)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func IconHandler(ctx *fasthttp.RequestCtx) {
	f, err := os.OpenFile(PathToView+"/favicon.ico", os.O_RDONLY, 0444)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(ctx, f)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}

	ctx.SetContentType(ContentTypeIcon)
}

func ImageHandler(ctx *fasthttp.RequestCtx) {
	f, err := os.OpenFile(PathToView+"/img/"+ctx.UserValue("imgFile").(string), os.O_RDONLY, 0444)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(ctx, f)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}

	ctx.SetContentType(ContentTypePng)
}

func CssHandler(ctx *fasthttp.RequestCtx) {
	f, err := os.OpenFile(PathToView+"/css/"+ctx.UserValue("cssFile").(string), os.O_RDONLY, 0444)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(ctx, f)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println(err)
		return
	}

	ctx.SetContentType(ContentTypeCss)
}
