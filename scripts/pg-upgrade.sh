#! /usr/bin/env bash
#
# This script will extract a new version of Packet Guardian from a distribution
# tarball and run the upgrade script included for the new version
#

# Check running as root
if [[ $UID -ne 0 ]]; then
    exec sudo "$0" "$@"
fi

# Copy to temporary location and run from there
# The upgrade script will edit this file while it's running
# which causes adverse side-effects
if [[ ! "`dirname $0`" =~ ^/tmp/.sh-tmp ]]; then
    mkdir -p /tmp/.sh-tmp/
    DIST="/tmp/.sh-tmp/$( basename $0 )"
    install -m 700 "$0" $DIST
    exec $DIST "$@"
else
    # Delete temporary copy
    rm "$0"
fi

ALL_YES=""
if [[ $1 == "-y" ]]; then
    ALL_YES="t"
    shift
fi

SRC_TAR="$(python -c "import os,sys; print os.path.realpath(sys.argv[1])" $1 2>/dev/null)"

if [[ ! -f $SRC_TAR ]]; then
    echo "Source tar file not found"
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

confirm "Is the file $SRC_TAR correct?"
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
