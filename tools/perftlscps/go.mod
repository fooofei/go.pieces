module github.com/fooofei/go_pieces/tools/perftlscps

go 1.21

replace github.com/fooofei/go_pieces/pkg => ../../pkg

require (
	github.com/bifurcation/mint v0.0.0-20210616192047-fd18df995463
	github.com/fooofei/go_pieces/pkg v0.0.0-20240904074352-b46acbef448f
)

require golang.org/x/crypto v0.31.0 // indirect
