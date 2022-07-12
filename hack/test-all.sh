#!/bin/bash

set -e -x -u

bash -c "while true; do echo \$(ps aux | wc -l); sleep 2; done" &
./hack/build.sh
./hack/test.sh
./hack/test-e2e.sh
./hack/test-examples.sh

echo ALL SUCCESS
