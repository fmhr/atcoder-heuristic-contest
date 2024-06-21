#!/bin/bash
set -xe

RAW_SEED=${1:-0000}
SEED=$(printf "%04d" $RAW_SEED)

make build

./bin/a.out < tools/in/${SEED}.txt > out/${SEED}.txt
(cd ./tools && cargo run -r --bin vis in/${SEED}.txt ../out/${SEED}.txt)