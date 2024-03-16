package main

import "github.com/goplus/yap"

func main2() {
	y := yap.New()
	y.GET("/", func(ctx *yap.Context) {
		ctx.TEXT(200, "text/html", `<html><body>Hello, YAP!</body></html>`)
	})
	y.GET("/p/:id", func(ctx *yap.Context) {
		ctx.JSON(200, yap.H{
			"id": ctx.Param("id"),
		})
	})
	y.Run("localhost:8080")
}
