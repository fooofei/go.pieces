GOROOT=/usr/local/Cellar/go/1.10/libexec #gosetup
GOPATH=/Users/hujianfei/go #gosetup
/usr/local/Cellar/go/1.10/libexec/bin/go build -o /private/var/folders/pz/wk429kjs11l_7dphvhtj9xl80000gn/T/___go_build_lession1_go -gcflags "all=-N -l" /Users/hujianfei/go/src/awesomeProject/lession1.go #gosetup
/Applications/GoLand.app/Contents/plugins/intellij-go-plugin/lib/dlv/mac/dlv --listen=localhost:61318 --headless=true --api-version=2 --backend=native exec /private/var/folders/pz/wk429kjs11l_7dphvhtj9xl80000gn/T/___go_build_lession1_go -- #gosetup
API server listening at: 127.0.0.1:61318



Goland 2018.2.2 Windows macOS 双平台确认把 delve 作为 debugger 了

20180829 进展

尝试做ssh remote debug，在端A linux 平台运行被调试程序 dlv --listen=:2345 --headless=true --api-version=2 exec ./demo.exe

这个程序使用的构建命令 go build -gcflags "all=-N -l" github.com/app/demo

在端B，通过命令行 dlv connect <ip_a>:2345 连接成功，进入命令行，还是看不到go代码。

在端C，这是app编译的所在机器，dlv connect <ip_a>:2345 连接成功能连接成功，能关联到源码。

goland 无法连接，VisualStudioCode连接成功，无法加断点，无法关联代码。