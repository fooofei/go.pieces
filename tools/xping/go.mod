module github.com/fooofei/go_pieces/tools/ping

go 1.18

replace (
	github.com/fooofei/go_pieces/pkg => ../../pkg
	github.com/kbinani/win => ./win
	github.com/montanaflynn/stats => ./stats
)

require (
	github.com/fooofei/go_pieces/pkg v0.0.0-00010101000000-000000000000
	github.com/montanaflynn/stats v0.6.6
	github.com/kbinani/win v0.3.0
)
