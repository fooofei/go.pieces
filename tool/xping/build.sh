#!/bin/bash

SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"

function compile() {
    name=$1
    echo "compiling $name"
    cd $SCRIPTPATH/$name
    rm -f $name.linux
    GOOS=linux GOARCH=amd64 go build -mod=vendor -v
    mv $name $name.linux
    go build -mod=vendor -v
    GOOS=windows GOARCH=amd64 go build -mod=vendor -v
}

compile httping
compile tcping
compile udping
