#!/bin/bash

set -ue
set -o pipefail

function main() {
  while IFS= read -r -d '' path; do
    pushd ${path}
    echo "in $(pwd)"
    if [[ -f go.mod ]] ; then
      go mod tidy || true
      go get -u ...
    fi
    popd
  done < <(find . -type d 2>/dev/null -print0)
}

main