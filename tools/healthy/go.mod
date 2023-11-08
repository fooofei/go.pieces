module github.com/fooofei/go_pieces/tools/healthy

go 1.21

require (
	github.com/fooofei/go_pieces/pkg v0.0.0-20230408021751-72c996e52f52
	golang.org/x/net v0.17.0
)

require golang.org/x/sys v0.13.0 // indirect

replace github.com/fooofei/go_pieces/pkg => ../../pkg
