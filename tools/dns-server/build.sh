#!/bin/bash
bin=dns-server

rm -f $bin
rm -f $bin.linux

GOOS=linux GOARCH=amd64 go build -mod=vendor -v
mv $bin $bin.linux
go build -mod=vendor -v
