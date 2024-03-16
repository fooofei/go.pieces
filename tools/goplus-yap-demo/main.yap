
import (
    "io"
    "fmt"
)

get "/", ctx => {
	ctx.html `<html><body>Hello, YA3P!</body></html>`
}
get "/p/:id", ctx => {
    var body = ctx.Body
	ctx.json {
		"id": ctx.param("id"),
		"body": body,
	}
}

post "/", ctx => {
    var body = ctx.Body
    var body2,_ = io.ReadAll(ctx.Body)
    ctx.json {
        "test": "a",
        "body": body,
        "body2": fmt.Sprintf("%s", body2),
    }
}

run "localhost:8082"
