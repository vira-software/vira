package main

import (
	"github.com/vira-software/vira/v1"
)

func main() {
	server := vira.New()
	server.UseHandler(vira.Default(true))

	router := vira.NewRouter()
	router.Get("/hello", func(ctx *vira.Context) error {
		return ctx.HTML(200, "<h1>Hello, Vira!</h1>")
	})
	router.Otherwise(func(ctx *vira.Context) error {
		return ctx.JSON(200, map[string]any{
			"Host":    ctx.Host,
			"Method":  ctx.Method,
			"Path":    ctx.Path,
			"URI":     ctx.Req.RequestURI,
			"Headers": ctx.Req.Header,
		})
	})

	server.UseHandler(router)
	server.Error(server.Listen(":3000"))
}
