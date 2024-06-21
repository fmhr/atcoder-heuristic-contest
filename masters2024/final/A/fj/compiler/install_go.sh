#!/bin/bash

cd /tmp                                    
wget -q https://go.dev/dl/go1.20.6.linux-amd64.tar.gz
tar -C /opt -xf go1.20.6.linux-amd64.tar.gz
export PATH=$PATH:/opt/go/bin # 一時的にパスを通す　コンテナには反映されない

mkdir -p /go/src/atcoder.jp/golang
cd /go/src/atcoder.jp/golang
go mod init atcoder.jp/golang
go get -u github.com/emirpasic/gods/...
go get -u gonum.org/v1/gonum/...
go get -u github.com/liyue201/gostl/...
go get -u golang.org/x/exp/