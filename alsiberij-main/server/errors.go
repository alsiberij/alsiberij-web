package server

import (
	"github.com/valyala/fasthttp"
	"html/template"
	"log"
)

type (
	ErrorTemplateData struct {
		ErrorCode    int
		ErrorMessage string
	}
)

func setHttpErrorHTML(ctx *fasthttp.RequestCtx, templateData ErrorTemplateData) {
	tmpl, err := template.ParseFiles(PathToView + "/html/error.html")
	if err != nil {
		log.Println(err)
		return
	}
	_ = tmpl.Execute(ctx, templateData)
	ctx.SetContentType(ContentTypeHtml)
	ctx.SetStatusCode(templateData.ErrorCode)
}

func Set404HTML(ctx *fasthttp.RequestCtx) {
	setHttpErrorHTML(ctx, ErrorTemplateData{
		ErrorCode:    fasthttp.StatusNotFound,
		ErrorMessage: "NOT FOUND",
	})
}

func Set405HTML(ctx *fasthttp.RequestCtx) {
	setHttpErrorHTML(ctx, ErrorTemplateData{
		ErrorCode:    fasthttp.StatusMethodNotAllowed,
		ErrorMessage: "METHOD NOT ALLOWED",
	})
}

func Set500HTML(ctx *fasthttp.RequestCtx, i interface{}) {
	setHttpErrorHTML(ctx, ErrorTemplateData{
		ErrorCode:    fasthttp.StatusInternalServerError,
		ErrorMessage: "INTERNAL SERVER ERROR",
	})
}
