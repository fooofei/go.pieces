module github.com/fooofei/go_pieces/tools/perftlscps

go 1.26.0

replace github.com/fooofei/go_pieces/pkg => ../../pkg

require (
	github.com/bifurcation/mint v0.0.0-20210616192047-fd18df995463
	github.com/fooofei/go_pieces/pkg v0.0.0-20260716033500-f136ba4e8456
)

require golang.org/x/crypto v0.58.0 // indirect
