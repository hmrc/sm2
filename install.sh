#!/bin/env bash

# defaults
INSTALL_DIR="/usr/local/bin"
CONFIG_LOC="${HOME}/.sm2/service-manager-config"

TARGET_OS="linux"
TARGET_ARCH="intel"
TARGET_VERSION="1.0.11"

if [[ "$(uname)" == "Darwin" ]]; then
    TARGET_OS="apple"
fi

if [[ "$(uname -p)" == "arm" ]]; then
    TARGET_ARCH="arm64"
fi

# util
line_break() {
  echo ""
}

download_sm2() {
  # https://unix.stackexchange.com/a/84980
  TEMPORARY_DIRECTORY=$(mktemp -d 2>/dev/null || mktemp -d -t 'sm2')
  line_break
  echo "Downloading sm2 v${TARGET_VERSION}..."

  # not using curl --output-dir as only added in curl 7.73.0 and would result in error if user has older version
  cd "${TEMPORARY_DIRECTORY}" && \
  curl -s -L -O "https://github.com/hmrc/sm2/releases/download/v${TARGET_VERSION}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" && \
  cd - > /dev/null && \
  unzip "${TEMPORARY_DIRECTORY}/sm2-${TARGET_VERSION}-${TARGET_OS}-${TARGET_ARCH}.zip" -d "${TEMPORARY_DIRECTORY}" && \
  chmod +x "${TEMPORARY_DIRECTORY}/sm2"
}

# handle Ctrl+C and exit
cleanup() {
  rm -rf "${TEMPORARY_DIRECTORY}"
  line_break
  echo "Script terminated by user."
  exit 1
}

# trap Ctrl+C and execute the cleanup function
trap cleanup INT

is_update() {
  # is it already on the $PATH
  if which sm2 >/dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

is_sudo_required() {
  local target_directory="$1"

  # Check if the user has write access to the directory
  if [ -w "$target_directory" ]; then
    return 1
  else
    return 0
  fi
}

prompt_yes_no() {
    while true; do
      read -p "$1 (y/n): " answer
        case $answer in
          [Yy]* ) return 0;;
          [Nn]* ) return 1;;
          * ) echo "Please answer [y]es or [n]o.";;
        esac
    done
}

choose_install_dir() {
  line_break
  echo "By default, sm2 will be installed to ${INSTALL_DIR} (recommended)"
  line_break

  if prompt_yes_no "Would you like to select a different location from your PATH?"; then
      # split the $PATH into an array of directories
      IFS=':' read -ra path_array <<< "$PATH"

      # Display numbered list of directories
      echo "Choose where sm2 should be installed from the following list:"
      line_break
      for ((i = 0; i < ${#path_array[@]}; i++)); do
        local directory="${path_array[i]}"
        if is_sudo_required "${directory}"; then
          echo "${i}) ${directory} (sudo required)"
        else
          echo "${i}) ${directory}"
        fi

        if [ "${directory}" == "${INSTALL_DIR}" ]; then
          local default="${i}"
        fi
      done

      line_break
      # shellcheck disable=SC2016
      echo 'These are all locations on your $PATH - if you want to install'
      # shellcheck disable=SC2016
      echo 'to a different location, then please update your $PATH and re-run'
      echo 'the script.'
      line_break

      # shellcheck disable=SC2162
      read -p "Enter the number of the directory you want to choose (default ${default}): " choice

      # validate
      if [ -z "$choice" ]; then
        # default selected
        return 0
      elif [[ "$choice" =~ ^[0-9]+$ && "$choice" -ge 0 && "$choice" -lt ${#path_array[@]} ]]; then
        INSTALL_DIR="${path_array[choice]}"
      else
        line_break
        echo "Invalid input - exiting..."
        exit 1
      fi
    else
      # happy with default
      return 0
    fi
}

config_exists() {
  if [ -d "${WORKSPACE}/service-manager-config" ] || [ -d "${HOME}/.sm2/service-manager-config" ]; then
    return 0
  else
    return 1
  fi
}

workspace_exists() {
  if [ -n "$WORKSPACE" ]; then
    return 0
  else
    return 1
  fi
}

choose_config_git() {
  line_break
  echo "Please specify the git repository that contains service-manager-config:"
  line_break
  # shellcheck disable=SC2162
  read -p "git clone " CONFIG_GIT
}

### main logic starts here ###

# check if this is an update or a fresh install
if is_update; then
  INSTALL_DIR=$(dirname "$(which sm2)")

  echo "Detected an existing sm2 installation in ${INSTALL_DIR} - updating..."
else
  # prompt user to choose a location on their $PATH
  choose_install_dir
fi

# download, extract and chmod sm2 binary
download_sm2

line_break
echo "Moving sm2 binary to ${INSTALL_DIR}..."

if is_sudo_required "${INSTALL_DIR}"; then
  line_break
  echo "sudo required, you may be prompted for your password..."
  sudo mv "${TEMPORARY_DIRECTORY}/sm2" "${INSTALL_DIR}/sm2"
else
  mv "${TEMPORARY_DIRECTORY}/sm2" "${INSTALL_DIR}/sm2"
fi

if ! config_exists; then
  line_break
  echo "sm2 requires a service-manager-config git repository..."

  if workspace_exists; then
    CONFIG_LOC="${WORKSPACE}/service-manager-config"
    line_break
    echo "Detected WORKSPACE environment variable, config will be cloned into ${CONFIG_LOC}..."
  else
    line_break
    echo "service-manager-config will be cloned into sm2's default location: ${CONFIG_LOC}..."
  fi
  choose_config_git
  git clone "${CONFIG_GIT}" "${CONFIG_LOC}"
fi

rm -rf "${TEMPORARY_DIRECTORY}"

echo "Successfully installed!"
line_break
echo "Running 'sm2 --diagnostic'..."
line_break

sm2 --diagnostic
