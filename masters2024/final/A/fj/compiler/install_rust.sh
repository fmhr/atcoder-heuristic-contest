#!/bin/bash

set -xe

apt-get update
apt-get install -y --no-install-recommends build-essential ca-certificates curl

rust_version=1.70.0

# Detect architecture and set the target variable accordingly.
case $(uname -m) in
    x86_64)   target="x86_64-unknown-linux-gnu" ;;
    aarch64)  target="aarch64-unknown-linux-gnu" ;;
    *)        echo "Unsupported architecture"; exit 1 ;;
esac

# Download and install Rust for the detected architecture.
curl "https://static.rust-lang.org/dist/rust-$rust_version-$target.tar.gz" -fO --output-dir /tmp
tar xvf "/tmp/rust-$rust_version-$target.tar.gz" -C /tmp
/tmp/rust-$rust_version-$target/install.sh

cargo -vV
[ "$(command -v cargo)" = /usr/local/bin/cargo ]
[ "$(cargo -vV | sed -n 's/release: \(.*\)/\1/p')" = "$rust_version" ]
[ "$(cargo -vV | sed -n 's/host: \(.*\)/\1/p')" = "$target" ]

mkdir ./.cargo ./src

cat > ./.cargo/config.toml << EOF
[build]
rustflags = [
    "--cfg", "atcoder",
]
EOF

cat > ./Cargo.toml << EOF
[profile.release]
lto = true

[package]
name = "main"
version = "0.0.0"
edition = "2021"
publish = false

# Dependencies omitted for brevity

EOF

# Fetching Cargo.lock from the repository
curl https://raw.githubusercontent.com/rust-lang-ja/atcoder-proposal/fe6aa6179d074d3a565d3c3db256db54071a38f9/Cargo.lock -fO

# Transitive dependencies license information is available in the repository files (omitted for brevity)

echo 'fn main() {}' > ./src/main.rs

cargo build -vv --release
rm target/release/main