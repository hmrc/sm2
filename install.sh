#!/bin/env bash

TARGET_OS="linux"
TARGET_ARCH="intel"
TARGET_VERSION="1.0.9"

if [[ "$(uname)" == "Darwin" ]]; then
    TARGET_OS="apple"
fi

if [[ "$(uname -p)" == "arm" ]]; then
    TARGET_ARCH="arm64"
fi

download_sm2() {
  # https://unix.stackexchange.com/a/84980
  TEMPORARY_DIRECTORY=$(mktemp -d 2>/dev/null || mktemp -d -t 'sm2')
  echo "Downloading sm2 v${TARGET_VERSION}..."

  curl -s -L -O "https://github.com/hmrc/sm2/releases/download/v${TARGET_VERSION}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" --output-dir "${TEMPORARY_DIRECTORY}" && \
    unzip "${TEMPORARY_DIRECTORY}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" -d "${TEMPORARY_DIRECTORY}" && \
    rm "${TEMPORARY_DIRECTORY}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" && \
    chmod +x "${TEMPORARY_DIRECTORY}/sm2"

  echo "Moving sm2 to /usr/local/bin..."
  sudo mv "${TEMPORARY_DIRECTORY}/sm2" /usr/local/bin/sm2
  echo "Successfully installed!"

  rm -fr "${TEMPORARY_DIRECTORY}"
}

create_workspace_folder() {
  DEFAULT_WORKSPACE_FOLDER="${HOME}/.sm2"
  # If the user has explicitly set a WORKSPACE location then respect that.
  WORKSPACE_FOLDER="${WORKSPACE:-"${DEFAULT_WORKSPACE_FOLDER}"}"
  echo "sm2 workspace folder is: ${WORKSPACE_FOLDER}"

  if [[ ! -d "${WORKSPACE_FOLDER}" ]]; then
    echo "Creating sm2 workspace folder in: ${WORKSPACE_FOLDER}"
    mkdir -p "${WORKSPACE_FOLDER}"
  fi
}

clone_service_manager_config() {
  if [[ -z "$1" ]]; then
      CONFIG_REPO="git@github.com:hmrc/service-manager-config.git"
  else
      CONFIG_REPO="$1"
  fi

  CONFIG_FOLDER="${WORKSPACE_FOLDER}/service-manager-config"
  if [[ ! -d "${CONFIG_FOLDER}" ]]; then
      echo "Cloning ${CONFIG_REPO} into ${CONFIG_FOLDER}"
      git clone "${CONFIG_REPO}" "${CONFIG_FOLDER}"
  else
    echo "Updating sm2 config in ${CONFIG_FOLDER}"
    # shellcheck disable=SC2164
    cd "${CONFIG_FOLDER}"
    git fetch --all --prune
    git pull --rebase
  fi
}

# Main
download_sm2
create_workspace_folder
clone_service_manager_config "${@}"

echo "Running sm2 --update..."

sm2 --update
