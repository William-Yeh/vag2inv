# building i386/x86_64 Windows/Linux/MacOSX executables via Dockerized Go compiler
#
# @see https://registry.hub.docker.com/_/golang/
#

# pull base image
FROM golang:1.5.1
MAINTAINER William Yeh <william.pjyeh@gmail.com>

ENV GOPATH    /opt 
WORKDIR       /opt


# fetch imported Go lib...
RUN  go get github.com/docopt/docopt-go
COPY vag2inv.go /opt/

# compile...
RUN  GOOS=windows GOARCH=386   \
     go build -o vag2inv-i386.exe

RUN  GOOS=windows GOARCH=amd64 \
     go build -o vag2inv-x86_64.exe

RUN  GOOS=linux   GOARCH=386    \
     go build -o vag2inv-linux-i386

RUN  GOOS=linux   GOARCH=amd64  \
     go build -o vag2inv-linux-x86_64

RUN  GOOS=darwin  GOARCH=amd64  \
     go build -o vag2inv-mac


# copy executable
RUN    mkdir -p /dist
VOLUME [ "/dist" ]
CMD    cp  vag2inv-*  /dist
