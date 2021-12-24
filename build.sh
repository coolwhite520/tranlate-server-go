#!/bin/sh
CGO_ENABLED=1 GOOS=linux  GOARCH=amd64  CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build -ldflags "-s -w"
#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build