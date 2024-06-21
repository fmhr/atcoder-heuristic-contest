#!/bin/bash

apt-get update
apt-get install -y g++-12
apt-get install -y libgmp3-dev

# /tmpディレクトリに移動してから作業を行う
cd /tmp

# ac-libraryのダウンロードとインストール
mkdir -p /opt/ac-library
wget https://github.com/atcoder/ac-library/releases/download/v1.5.1/ac-library.zip -O ac-library.zip
unzip /tmp/ac-library.zip -d /opt/ac-library

# boostのダウンロードとインストール
apt-get install -y build-essential
# wget https://boostorg.jfrog.io/artifactory/main/release/1.82.0/source/boost_1_82_0.tar.gz -O boost_1_82_0.tar.gz
wget https://github.com/boostorg/boost/releases/download/boost-1.82.0/boost-1.82.0.tar.gz -O boost_1_82_0.tar.gz
tar xf boost_1_82_0.tar.gz
cd boost-1.82.0
./bootstrap.sh --with-toolset=gcc --without-libraries=mpi,graph_parallel
./b2 -j3 toolset=gcc variant=release link=static runtime-link=static cxxflags="-std=c++20" stage
./b2 -j3 toolset=gcc variant=release link=static runtime-link=static cxxflags="-std=c++20" --prefix=/opt/boost/gcc install

# Eigenのインストール
apt-get install -y libeigen3-dev=3.4.0-2ubuntu2
