#!/bin/bash

set -euo pipefail

dst_dir="${CARVEL_INSTALL_BIN_DIR:-${K14SIO_INSTALL_BIN_DIR:-/usr/local/bin}}"
go run ./hack/deps.go -dest="${dst_dir}" -dev
