cur=$(dirname "$(readlink -f "$0")")
go build -v .
ssh -o "ProxyCommand $cur/client 127.0.0.1 3389 %h %p" $@