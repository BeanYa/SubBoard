#!/bin/bash
export PATH=$HOME/go/bin:$PATH
export GOPATH=$HOME/gopath
export GOPROXY=https://goproxy.cn,direct
mkdir -p $GOPATH
cd /mnt/c/universe/workspace/repo/subboard/backend
go env GOROOT GOPROXY CGO_ENABLED
echo "---"
go mod tidy
echo "go mod tidy exit code: $?"
go build -o submanager .
echo "Build exit code: $?"