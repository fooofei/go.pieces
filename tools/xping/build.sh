#!/bin/bash

set -e
set -x

var_cur="$( cd "$(dirname "$0")" ; pwd -P )"

function compile_target() {
    var_name=$1
    echo "--> compiling ${var_cur}/cmd/${var_name}"
    pushd "${var_cur}/cmd/${var_name}"
    go mod vendor
    GOOS=linux GOARCH=amd64 go build -mod=vendor -tags netgo -v -o ${var_name}-linux-amd64
    GOOS=linux GOARCH=arm64 go build -mod=vendor -tags netgo -v -o ${var_name}-linux-arm64
    GOOS=windows GOARCH=amd64 go build -mod=vendor -tags netgo -v -o ${var_name}-windows.exe
    GOOS=darwin GOARCH=amd64 go build -mod=vendor -tags netgo -v -o ${var_name}-darwin-amd64
    popd
}

# clone the remote repo to local, for the remote repo not have go.mod
function get_kbinani_win() {
  pushd "${var_cur}"
  git clone https://github.com/kbinani/win.git
  cat << EOF > win/go.mod
module github.com/kbinani/win

go 1.19
EOF
  popd
}

# clone the remote repo to local, for the remote repo not have go.mod
function get_montanaflynn_stats() {
  pushd "${var_cur}"
  git clone https://github.com/montanaflynn/stats.git
  popd
}

get_kbinani_win || true
get_montanaflynn_stats || true

compile_target http
compile_target tcp
compile_target udp


