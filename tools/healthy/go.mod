module github.com/fooofei/go_pieces/tools/healthy

go 1.21

require (
	github.com/fooofei/go_pieces/pkg v0.0.0-20231108115809-fc3d9ecc3762
	golang.org/x/net v0.17.0
)

require golang.org/x/sys v0.15.0 // indirect

replace github.com/fooofei/go_pieces/pkg => ../../pkg
