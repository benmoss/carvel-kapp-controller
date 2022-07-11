#!/bin/bash

set -e -x -u

./hack/build.sh
ulimit -u
./hack/test.sh
ulimit -u
./hack/test-e2e.sh
ulimit -u
./hack/test-examples.sh
ulimit -u

echo ALL SUCCESS
