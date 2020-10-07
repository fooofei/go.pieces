#!/bin/bash

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"

function compile() {
    name=$1
    echo "compiling $name"
    cd $SCRIPTPATH/$name
    GOOS=linux GOARCH=amd64 go build -mod=vendor -tags netgo -v -o $name.linux.amd64
    GOOS=linux GOARCH=arm64 go build -mod=vendor -tags netgo -v -o $name.linux.arm64
    GOOS=windows GOARCH=amd64 go build -mod=vendor -tags netgo -v -o $name.windows.exe
    GOOS=darwin GOARCH=amd64 go build -mod=vendor -tags netgo -v -o $name.darwin.amd64
}

compile httping
compile tcping
compile udping
