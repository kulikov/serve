#!/bin/bash
set -ex

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

pushd $DIR

for os in linux darwin; do
  GOOS=$os CGO_ENABLED=0 go build -v -o dist/serve-$os apps/serve.go
done

popd
