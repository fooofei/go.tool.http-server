#!/usr/bin/env bash
cur=$(dirname "$(readlink -f $0)")

bin=sim_http_server

GOOS=linux GOARCH=arm64 go build -mod=vendor -v -o $bin.linux.arm64 $cur/cmd/server
GOOS=linux GOARCH=amd64 go build -mod=vendor -v -o $bin.linux.amd64 $cur/cmd/server
GOOS=windows GOARCH=amd64 go build -mod=vendor -v -o $bin.exe $cur/cmd/server
