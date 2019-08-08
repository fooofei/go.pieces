#!/usr/bin/env bash
bin=publicip
GOOS=linux GOARCH=amd64 go build -mod=vendor -v
mv $bin $bin.linux

go build -mod=vendor -v