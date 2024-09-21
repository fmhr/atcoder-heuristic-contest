# ベースイメージとしてUbuntu 24.04を使用
FROM ubuntu:24.04 AS builder

# 環境変数の設定
ENV GO_VERSION=1.20.13 \
    PATH="/usr/local/go/bin:/root/.cargo/bin:${PATH}" \
    DEBIAN_FRONTEND=noninteractive

# 必要なパッケージをインストール
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    build-essential \
    ca-certificates \
    git \
    language-pack-ja \
    && rm -rf /var/lib/apt/lists/*

# Goのセットアップ（アーキテクチャに応じて適切なバイナリを選択）
RUN case $(uname -m) in \
        x86_64) ARCH=amd64 ;; \
        aarch64|arm64) ARCH=arm64 ;; \
        *) echo "Unsupported architecture"; exit 1 ;; \
    esac && \
    curl -LO https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz && \
    rm go${GO_VERSION}.linux-${ARCH}.tar.gz

# Rustのセットアップ
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y 

# ロケール設定（日本語に設定）
RUN locale-gen ja_JP.UTF-8 && \
    update-locale LANG=ja_JP.UTF-8

# 環境変数の設定
ENV LANG=ja_JP.UTF-8
ENV LANGUAGE=ja_JP:ja
ENV LC_ALL=ja_JP.UTF-8

# イメージのエントリポイント（bashをデフォルトで起動）
CMD ["/bin/bash"]
