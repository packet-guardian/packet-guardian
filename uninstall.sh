#! /usr/bin/env bash

PRG_DIR="/opt/packet-guardian"
SERVICE_FILE="/etc/init/pg.conf"
LOG_DIR="/var/log/packet-guardian"
DATA_DIR="/var/lib/packet-guardian"
CONFIG_DIR="/etc/packet-guardian"
APPARMOR_CONFIG_FILE="/etc/apparmor.d/opt.packet-guardian.bin.pg"

# Check running as root
if [[ $UID -ne 0 ]]; then
    echo "This file must be ran as root."
    exit 1
fi

echo -n "Are you sure you want to uninstall Packet Guardian? [y/N]: "
read -n 1 imsure
echo

if [[ $imsure != "y" ]]; then
    exit 0
fi

rm -rf $PRG_DIR

if [[ $1 -eq "-p" ]]; then
    echo -n "Are you sure you want to purge Packet Guardian config files? [y/N]: "
    read -n 1 imsure
    echo

    if [[ $imsure != "y" ]]; then
        exit 0
    fi

    service pg stop
    rm -rf $SERVICE_FILE
    rm -rf $LOG_DIR
    rm -rf $DATA_DIR
    rm -rf $CONFIG_DIR
    rm -rf $APPARMOR_CONFIG_FILE
fi
