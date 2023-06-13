module github.com/fooofei/go_pieces/tools/ping

go 1.21

replace (
	github.com/fooofei/go_pieces/pkg => ../../pkg
	github.com/kbinani/win => ./win
	github.com/montanaflynn/stats => ./stats
)

require (
	github.com/fooofei/go_pieces/pkg v0.0.0-20230408021751-72c996e52f52
	github.com/kbinani/win v0.3.0
	github.com/montanaflynn/stats v0.7.1
)
