#!/bin/bash

set -o errexit
set -o nounset

for size in 1 200 10000; do
    echo "Size ${size}"
    diff <(go run pricer.go ${size} < pricer.in) <(gzip -dc pricer.out.${size}.gz)
done
