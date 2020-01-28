#!/bin/bash -eux

export GOPATH=$(pwd)

pushd $GOPATH/dp-api-router
  make test
popd
