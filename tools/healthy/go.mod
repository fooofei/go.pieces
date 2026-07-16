module github.com/fooofei/go_pieces/tools/healthy

go 1.26.0

require (
	github.com/fooofei/go_pieces/pkg v0.0.0-20260716033500-f136ba4e8456
	golang.org/x/net v0.62.0
)

require golang.org/x/sys v0.50.0 // indirect

replace github.com/fooofei/go_pieces/pkg => ../../pkg
