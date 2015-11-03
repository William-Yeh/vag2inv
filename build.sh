#!/bin/bash
#
# scirpt for compiling go source and generating i386/x86_64 Windows/Linux/MacOSX executables.
#

set -e
set -x

go get github.com/docopt/docopt-go

mkdir -p dist

GOOS=windows GOARCH=386    go build -o dist/vag2inv-i386.exe
GOOS=windows GOARCH=amd64  go build -o dist/vag2inv-x86_64.exe

GOOS=linux   GOARCH=386    go build -o dist/vag2inv-linux-i386
GOOS=linux   GOARCH=amd64  go build -o dist/vag2inv-linux-x86_64

GOOS=darwin  GOARCH=amd64  go build -o dist/vag2inv-mac
