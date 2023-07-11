#!/bin/bash

TARGET_OS="linux"
TARGET_ARCH="intel"
TARGET_VERSION="1.0.9"

if [[ "$(uname)" == "Darwin" ]]; then
    TARGET_OS="apple"
fi

if [[ "$(uname -p)" == "arm" ]]; then
    TARGET_ARCH="arm64"
fi

cd &&
curl -s -L -O "https://github.com/hmrc/sm2/releases/download/v$TARGET_VERSION/sm2-$TARGET_VERSION-$TARGET_OS-$TARGET_ARCH.zip" &&
unzip sm2-$TARGET_VERSION-$TARGET_OS-$TARGET_ARCH.zip && rm sm2-$TARGET_VERSION-$TARGET_OS-$TARGET_ARCH.zip &&
chmod +x sm2
