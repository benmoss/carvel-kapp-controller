#!/bin/bash

set -e

docker build -t ko.local/kc-base -q .
ytt -f config | ko resolve -f- | kapp deploy -a kc -f- -c -y

source ./hack/secretgen-controller.sh
deploy_secretgen-controller
