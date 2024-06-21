FROM rust:buster as toolsbuilder
WORKDIR /work

# 直接ダウンロードするとき　例)AHC027
#RUN wget -q https://img.atcoder.jp/ahc027/aPdjCUIZ_v2.zip -O tools.zip \
#    && unzip tools.zip 

# ローカルのtoolsをコピーするときは tools/　をDockerfileと同じ階層に置く
COPY ./tools /work/tools
RUN (cd tools && cargo build --release --quiet)


FROM golang:1.21 as workerbuild
WORKDIR /work
RUN git clone https://github.com/fmhr/fj.git

WORKDIR /work/fj
RUN go mod tidy

WORKDIR /work/fj/cmd/worker 
RUN CGO_ENABLED=0 go build -o /work/worker


# 3. 実行時の環境
FROM ubuntu:latest
WORKDIR /judge

RUN apt-get update && apt-get install ca-certificates openssl

COPY --from=workerbuild /work/worker /judge/worker
COPY --from=toolsbuilder /work/tools/target/release/* /judge/tools/target/release/

RUN mkdir -p /judge/bin

EXPOSE $PORT
CMD ["/judge/worker"]