#!/bin/sh
# 这行命令的使用，是因为使用了sqlite3
#CGO_ENABLED=1 GOOS=linux  GOARCH=amd64  CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build -ldflags "-s -w"
# 现在已经不使用sqlite3 ,而采用的mysql docker容器，所以下面的编译命令就可以了
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64  go build
#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build