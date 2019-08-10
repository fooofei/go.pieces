#!/usr/bin/env bash
bin=perftlscps
GOOS=linux GOARCH=amd64 go build -mod=vendor -v
mv $bin $bin.linux