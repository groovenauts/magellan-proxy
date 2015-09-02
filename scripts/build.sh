#!/bin/sh

export GOOS=linux
export GOARCH=amd64
gom build -o magellan-proxy_linux-amd64 github.com/groovenauts/magellan-proxy

