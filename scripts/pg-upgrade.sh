#! /usr/bin/env bash
#
# This script will extract a new version of Packet Guardian from a distribution
# tarball and run the upgrade script included for the new version
#

# Check running as root
if [[ $UID -ne 0 ]]; then
    echo "This file must be ran as root."
    exit 1
fi

ALL_YES=""
if [[ $1 == "-y" ]]; then
    ALL_YES="t"
    shift
fi

SRC_TAR="$(realpath $1)"

if [[ ! -f $SRC_TAR ]]; then
    echo "$SRC_TAR not found"
    exit 1
fi

SYSTEMD=""
if which systemctl >/dev/null 2>&1; then
    SYSTEMD="t"
fi

confirm() {
    if [[ -n $ALL_YES ]]; then
        return
    fi
    echo -n "$1 [y/N]: "
    read -n 1 imsure
    echo
    if [[ $imsure != "y" ]]; then
        exit 0
    fi
}

stopServices() {
    echo "Stopping Packet Guardian"
    if [[ -n $SYSTEMD ]]; then
        systemctl stop pg-dhcp >/dev/null 2>&1
        systemctl stop pg >/dev/null 2>&1
    else
        service pg-dhcp stop >/dev/null 2>&1
        service pg stop >/dev/null 2>&1
    fi
}

startServices() {
    echo "Starting Packet Guardian"
    if [[ -n $SYSTEMD ]]; then
        systemctl start pg-dhcp
        systemctl start pg
    else
        service pg-dhcp start
        service pg start
    fi
}

confirm "This will upgrade Packet Guardian, are you sure?"
cd /opt
stopServices
# Remove any old versions
rm -rf packet-guardian.old
# Move current version to old version
mv packet-guardian packet-guardian.old
# Extract tarball
tar -xzf $SRC_TAR

cd packet-guardian
# Run version specific upgrade script
./scripts/upgrade.sh -y
startServices
