GOROOT=/usr/local/Cellar/go/1.10/libexec #gosetup
GOPATH=/Users/hujianfei/go #gosetup
/usr/local/Cellar/go/1.10/libexec/bin/go build -o /private/var/folders/pz/wk429kjs11l_7dphvhtj9xl80000gn/T/___go_build_lession1_go -gcflags "all=-N -l" /Users/hujianfei/go/src/awesomeProject/lession1.go #gosetup
/Applications/GoLand.app/Contents/plugins/intellij-go-plugin/lib/dlv/mac/dlv --listen=localhost:61318 --headless=true --api-version=2 --backend=native exec /private/var/folders/pz/wk429kjs11l_7dphvhtj9xl80000gn/T/___go_build_lession1_go -- #gosetup
API server listening at: 127.0.0.1:61318
