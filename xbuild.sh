#!/bin/sh
# Generates cross builds for all supported platforms.
#
# This script is used to build binaries for all supported platforms. Cgo is
# disabled to make sure binaries are statically linked. Appropriate flags are
# given to the go compiler to strip binaries. These are then compressed in an
# archive form (`.zip` for windows and `.tar.gz` for the rest) within a folder
# named `dist`.

set -o verbose

mkdir -p dist

CGO_ENABLED=0 GOOS=darwin    GOARCH=386      go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-darwin-386.tar.gz      zed 
CGO_ENABLED=0 GOOS=darwin    GOARCH=amd64    go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-darwin-amd64.tar.gz    zed 
CGO_ENABLED=0 GOOS=freebsd   GOARCH=386      go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-freebsd-386.tar.gz     zed 
CGO_ENABLED=0 GOOS=freebsd   GOARCH=amd64    go build -ldflags='-s -w' && sync && tar czf dist/zed-freebsd-amd64.tar.gz   zed 
CGO_ENABLED=0 GOOS=freebsd   GOARCH=arm      go build -ldflags='-s -w' && sync && tar czf dist/zed-freebsd-arm.tar.gz     zed 
CGO_ENABLED=0 GOOS=linux     GOARCH=386      go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-linux-386.tar.gz       zed 
CGO_ENABLED=0 GOOS=linux     GOARCH=amd64    go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-linux-amd64.tar.gz     zed 
CGO_ENABLED=0 GOOS=linux     GOARCH=arm      go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-linux-arm.tar.gz       zed 
CGO_ENABLED=0 GOOS=linux     GOARCH=arm64    go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-linux-arm64.tar.gz     zed 
CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64    go build -ldflags='-s -w' && sync && tar czf dist/zed-linux-ppc64.tar.gz     zed 
CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64le  go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-linux-ppc64le.tar.gz   zed 
CGO_ENABLED=0 GOOS=netbsd    GOARCH=386      go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-netbsd-386.tar.gz      zed 
CGO_ENABLED=0 GOOS=netbsd    GOARCH=amd64    go build -ldflags='-s -w' && sync && tar czf dist/zed-netbsd-amd64.tar.gz    zed 
CGO_ENABLED=0 GOOS=netbsd    GOARCH=arm      go build -ldflags='-s -w' && sync && tar czf dist/zed-netbsd-arm.tar.gz      zed 
CGO_ENABLED=0 GOOS=openbsd   GOARCH=386      go build -ldflags='-s -w' && upx zed > /dev/null && sync && tar czf dist/zed-openbsd-386.tar.gz     zed 
CGO_ENABLED=0 GOOS=openbsd   GOARCH=amd64    go build -ldflags='-s -w' && sync && tar czf dist/zed-openbsd-amd64.tar.gz   zed 
CGO_ENABLED=0 GOOS=openbsd   GOARCH=arm      go build -ldflags='-s -w' && sync && tar czf dist/zed-openbsd-arm.tar.gz     zed 

CGO_ENABLED=0 GOOS=windows GOARCH=386   go build -ldflags='-s -w' && upx zed.exe > /dev/null.exe && sync && zip dist/zed-windows-386.zip   zed.exe 
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' && upx zed.exe > /dev/null.exe && sync && zip dist/zed-windows-amd64.zip zed.exe 
