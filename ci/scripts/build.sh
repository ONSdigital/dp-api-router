#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd

pushd $GOPATH/dp-api-router
  make build && mv build/$(go env GOOS)-$(go env GOARCH)/* $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
