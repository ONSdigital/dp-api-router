#!/bin/bash -eux

pushd dp-api-router
  make test-component
popd
