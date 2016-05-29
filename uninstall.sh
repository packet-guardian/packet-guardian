#! /usr/bin/env bash

# Check running as root
if [[ $UID -ne 0 ]]; then
    echo "This file must be ran as root."
    exit 1
fi

APP_DIR="/opt/packet-guardian"
UPSTART_SERVICE_DIR="/etc/init"
SYSTEMD_SERVICE_DIR="/etc/systemd/system"
LOG_DIR="/var/log/packet-guardian"
DATA_DIR="/var/lib/packet-guardian"
CONFIG_DIR="/etc/packet-guardian"
APPARMOR_DIR="/etc/apparmor.d"

SYSTEMD=""
APPARMOR_INSTALLED=""

if which systemctl >/dev/null 2>&1; then
    SYSTEMD="t"
fi
if which apparmor_status >/dev/null 2>&1; then
    APPARMOR_INSTALLED="t"
fi

confirm() {
    echo -n "$1 [y/N]: "
    read -n 1 imsure
    echo
    if [[ $imsure != "y" ]]; then
        exit 0
    fi
}

stopService() {
    echo "Stopping any running instances"
    if [[ -n $SYSTEMD ]]; then
        systemctl stop pg-dhcp >/dev/null 2>&1
        systemctl stop pg >/dev/null 2>&1
        systemctl disable pg-dhcp
        systemctl disable pg
    else
        service pg-dhcp stop >/dev/null 2>&1
        service pg stop >/dev/null 2>&1
    fi
}

uninstallService() {
    if [[ -n $SYSTEMD ]]; then
        echo "Uninstalling Systemd Service"
        rm -rf $SYSTEMD_SERVICE_DIR/pg.service
        rm -rf $SYSTEMD_SERVICE_DIR/pg-dhcp.service
    else
        echo "Uninstalling Upstart Service"
        rm -rf $UPSTART_SERVICE_DIR/pg.conf
        rm -rf $UPSTART_SERVICE_DIR/pg-dhcp.conf
    fi
}

uninstallAppArmorProfile() {
    # Install apparmor profile if available
    if [[ -n $APPARMOR_INSTALLED ]]; then
        echo "Uninstalling AppArmor Profile"
        rm -rf $APPARMOR_DIR/opt.packet-guardian.bin.pg
        rm -rf $APPARMOR_DIR/opt.packet-guardian.bin.dhcp
    fi
}

deleteApplicationFiles() {
    echo "Removing application files"
    rm -rf $APP_DIR
}

purgeConfigFiles() {
    confirm "Are you sure you want to purge Packet Guardian config files?"
    echo "Removing configuration and data directories"
    uninstallService
    uninstallAppArmorProfile
    rm -rf $LOG_DIR
    rm -rf $DATA_DIR
    rm -rf $CONFIG_DIR
    echo "Removing packetg user"
    userdel -f packetg
}

confirm "Are you sure you want to uninstall Packet Guardian?"
stopService
deleteApplicationFiles

if [[ $1 = "-p" ]]; then
    purgeConfigFiles
fi
