#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-api-router
  make audit
popd