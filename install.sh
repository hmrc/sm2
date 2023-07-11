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

echo "Downloading sm2 v$TARGET_VERSION..."

cd &&
curl -s -L -O "https://github.com/hmrc/sm2/releases/download/v$TARGET_VERSION/sm2-$TARGET_VERSION-$TARGET_OS-$TARGET_ARCH.zip" &&
unzip sm2-$TARGET_VERSION-$TARGET_OS-$TARGET_ARCH.zip && rm sm2-$TARGET_VERSION-$TARGET_OS-$TARGET_ARCH.zip &&
chmod +x sm2

echo "Moving sm2 to /usr/local/bin..."

sudo mv ~/sm2 /usr/local/bin/sm2

echo "Successfully installed!"

sm2 --version

echo "Running sm2 --update..."

sm2 --update

if [[ -z "${WORKSPACE}" ]]; then
    # shellcheck disable=SC2016
    echo 'Note: Your $WORKSPACE environment variable is not set. Instructions can be found in the user guide:'
    echo 'https://github.com/hmrc/sm2/blob/main/USERGUIDE.md#setup'
fi
