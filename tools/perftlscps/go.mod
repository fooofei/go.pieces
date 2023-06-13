module github.com/fooofei/go_pieces/tools/perftlscps

go 1.21

replace github.com/fooofei/go_pieces/pkg => ../../pkg

require (
	github.com/bifurcation/mint v0.0.0-20210616192047-fd18df995463
	github.com/fooofei/go_pieces/pkg v0.0.0-20230408021751-72c996e52f52
	github.com/pkg/errors v0.8.1
)

require golang.org/x/crypto v0.9.0 // indirect
