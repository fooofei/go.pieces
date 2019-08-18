cur=$(dirname "$(readlink -f "$0")")
go build -v .
ssh -o "ProxyCommand $cur/client 127.0.01 3389 %h %p" $@