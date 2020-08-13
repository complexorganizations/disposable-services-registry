#!/bin/bash

# Pre-Checks
function check-system-requirements() {
    # System requirements (go)
    if ! [ -x "$(command -v go)" ]; then
        echo "Error: go is not installed, please install go." >&2
        exit
    fi
    # System requirements (git)
    if ! [ -x "$(command -v git)" ]; then
        echo "Error: git is not installed, please install git." >&2
        exit
    fi
}

# Run the function and check for requirements
check-system-requirements

# Detect Operating System
function dist-check() {
    # shellcheck disable=SC1090
    if [ -e /etc/os-release ]; then
        # shellcheck disable=SC1091
        source /etc/os-release
        DISTRO=$ID
    fi
}

# Check Operating System
dist-check

function update() {
    # shellcheck disable=SC2233
    if ([ "$DISTRO" == "ubuntu" ] || [ "$DISTRO" == "debian" ]); then
        git clean -f && git reset --hard && git checkout master && git pull
        go run main.go
        echo "Please make a pull request."
    fi
}

# Run the function
update
