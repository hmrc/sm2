#!/bin/env bash
set -euo pipefail
IFS=$'\n\t'

TARGET_OS="linux"
TARGET_ARCH="intel"
TARGET_VERSION="2.2.0"

if [[ "$(uname)" == "Darwin" ]]; then
    TARGET_OS="apple"
fi

if [[ "$(uname -p)" == "arm" ]]; then
    TARGET_ARCH="arm64"
fi

validate_args() {
  if [[ -z "${1}" ]]; then
      echo "Error: sm2-configuration-repository-specification parameter not provided."
      echo
      usage
      exit 1
  fi
}

download_sm2() {
  # https://unix.stackexchange.com/a/84980
  TEMPORARY_DIRECTORY=$(mktemp -d 2>/dev/null || mktemp -d -t 'sm2')
  echo "Downloading sm2 v${TARGET_VERSION}..."

  # shellcheck disable=SC2164
  cd "${TEMPORARY_DIRECTORY}" && \
    curl -s -L -O "https://github.com/hmrc/sm2/releases/download/v${TARGET_VERSION}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" && \
    cd - && \
    unzip "${TEMPORARY_DIRECTORY}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" -d "${TEMPORARY_DIRECTORY}" && \
    rm "${TEMPORARY_DIRECTORY}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" && \
    chmod +x "${TEMPORARY_DIRECTORY}/sm2"

  # If the user already has sm2 on their path, even in a non-standard location, then use this location.
  if which sm2 >/dev/null 2>&1; then
      SM2_BINARY_LOCATION="$(which sm2)"
      echo "Found existing sm2 binary at ${SM2_BINARY_LOCATION}. Using existing location."
  else
      SM2_BINARY_LOCATION="/usr/local/bin/sm2"
      if [ ! -d "/usr/local/bin" ]; then
          sudo mkdir -p /usr/local/bin
      fi
  fi

  echo "Installing sm2 binary to ${SM2_BINARY_LOCATION}"
  sudo mv "${TEMPORARY_DIRECTORY}/sm2" "${SM2_BINARY_LOCATION}"
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
  CONFIG_REPO="${1}"

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

usage() {
  cat << EOF
  NAME
    sm2 installer

  USAGE
    install.sh [sm2-configuration-repository-specification]

    Example:
    install.sh "git@github.com:myorg/sm2-config.git"

  PURPOSE
    Installs Service Manager 2 (sm2).
    Obtains the necessary SM2 configuration as indicated by [sm2-configuration-repository-specification].
    Runs SM2 diagnostic utility, to verify correctness of sm2 installation.
EOF
}

# Main
validate_args "${@}"
download_sm2
create_workspace_folder
clone_service_manager_config "${@}"

echo "Running sm2 --diagnostic"

sm2 --diagnostic
