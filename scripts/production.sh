#!/bin/bash
# https://github.com/complexorganizations/disposable-services-registry

# Require script to be run as root
function super-user-check() {
  if [ "$EUID" -ne 0 ]; then
    echo "You need to run this script as super user."
    exit
  fi
}

# Check for root
super-user-check

# Detect Operating System
function dist-check() {
  if [ -e /etc/os-release ]; then
    # shellcheck disable=SC1091
    source /etc/os-release
    DISTRO=$ID
  fi
}

# Check Operating System
dist-check

# Pre-Checks system requirements
function installing-system-requirements() {
  if { [ "$DISTRO" == "ubuntu" ] || [ "$DISTRO" == "debian" ] || [ "$DISTRO" == "raspbian" ] || [ "$DISTRO" == "pop" ] || [ "$DISTRO" == "kali" ] || [ "$DISTRO" == "linuxmint" ] || [ "$DISTRO" == "fedora" ] || [ "$DISTRO" == "centos" ] || [ "$DISTRO" == "rhel" ] || [ "$DISTRO" == "arch" ] || [ "$DISTRO" == "manjaro" ] || [ "$DISTRO" == "alpine" ] || [ "$DISTRO" == "freebsd" ]; }; then
    if { [ ! -x "$(command -v curl)" ] || [ ! -x "$(command -v git)" ] || [ ! -x "$(command -v go)" ]; }; then
      if { [ "$DISTRO" == "ubuntu" ] || [ "$DISTRO" == "debian" ] || [ "$DISTRO" == "raspbian" ] || [ "$DISTRO" == "pop" ] || [ "$DISTRO" == "kali" ] || [ "$DISTRO" == "linuxmint" ]; }; then
        apt-get update && apt-get install curl git golang -y
      elif { [ "$DISTRO" == "fedora" ] || [ "$DISTRO" == "centos" ] || [ "$DISTRO" == "rhel" ]; }; then
        yum update -y && yum install curl git golang -y
      elif { [ "$DISTRO" == "arch" ] || [ "$DISTRO" == "manjaro" ]; }; then
        pacman -Syu --noconfirm curl git go
      elif [ "$DISTRO" == "alpine" ]; then
        apk update && apk add curl git golang
      elif [ "$DISTRO" == "freebsd" ]; then
        pkg update && pkg install curl git golang
      fi
    fi
  else
    echo "Error: $DISTRO not supported."
    exit
  fi
}

# Run the function and check for requirements
installing-system-requirements

function auto-update-every-day() {
  if [ ! -f "$GLOBAL_VARIABLES" ]; then
    echo "[Unit]
Description= Disposable Services Registry
After=network.target

[Service]
Type=oneshot
ExecStart=/

[Install]
WantedBy=multi-user.target" >> $GLOBAL_VARIABLES
  fi
}
