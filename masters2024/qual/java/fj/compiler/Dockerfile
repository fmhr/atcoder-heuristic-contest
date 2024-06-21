FROM golang:1.21 as serverbuild

WORKDIR /app

RUN git clone https://github.com/fmhr/fj.git \
    && cd fj \
    && go mod tidy

WORKDIR /app/fj/cmd/compiler

RUN go build -o /app/server



FROM ubuntu:latest

WORKDIR /work

RUN apt-get update && apt-get install -y \
    apt-utils \
    wget \
    unzip \
    gzip

COPY . /work
###################################################################
# 必要な言語のコメントアウトをはずしてね
###################################################################
# RUN sh install_go.sh 
# ENV PATH $PATH:/opt/go/bin
# RUN sh install_cpp20.sh
# RUN sh install_cs.sh
RUN sh install_java.sh 

COPY --from=serverbuild /app/server /work

RUN chmod +x /work/server

EXPOSE $PORT

CMD ["/work/server"]