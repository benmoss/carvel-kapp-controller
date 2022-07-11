#!/bin/bash

set -e -x -u

./hack/build.sh
echo ulimit: $(ulimit -u)
./hack/test.sh
echo ulimit: $(ulimit -u)
./hack/test-e2e.sh
echo ulimit: $(ulimit -u)
./hack/test-examples.sh
echo ulimit: $(ulimit -u)

echo ALL SUCCESS
