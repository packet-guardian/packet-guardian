#! /usr/bin/env bash

SRC_TAR="$1"
DST_DIR="/opt/packet-guardian"
SERVICE_FILE="/etc/init/pg.conf"
LOG_DIR="/var/log/packet-guardian"
DATA_DIR="/var/lib/packet-guardian"
CONFIG_DIR="/etc/packet-guardian"
APPARMOR_CONFIG_FILE="/etc/apparmor.d/opt.packet-guardian.bin.pg"
EXISTING=""

# Check running as root
if [[ $UID -ne 0 ]]; then
    echo "This file must be ran as root."
    exit 1
fi

# Check source file and destination directory
if [[ ! -f $SRC_TAR ]]; then
    echo "The source tarball doesn't appear to exist, please check your path and try again"
    exit 1
fi

SRC_TAR="$(realpath $SRC_TAR)"

if [[ ! -d $DST_DIR ]]; then
    mkdir $DST_DIR
else
    EXISTING="1"
fi

# Stop service if exists and running
if [[ -f $SERVICE_FILE ]]; then
    if service --status-all | grep -Fq 'pg'; then
        service pg stop
    fi
fi

# Save a copy of the old
if [[ -n $EXISTING ]]; then
    cp -R $DST_DIR $DST_DIR.old
fi

# Make user if doesn't exist
if [[ ! $(id -u packetg > /dev/null 2>&1) ]]; then
    useradd -M packetg
fi

# Extract binary and friends
cd /opt
tar -xzf $SRC_TAR
chown -R packetg:packetg $DST_DIR

# Install Upstart service
cd $DST_DIR
cp config/upstart.conf $SERVICE_FILE
chown root:root $SERVICE_FILE

# Set kernel permissions
setcap 'cap_net_bind_service=+ep' $DST_DIR/bin/pg

# Install apparmor profile if available
if [[ $(which apparmor_status) ]]; then
    if [[ $(which aa-complain) ]]; then
        cp config/apparmor.conf $APPARMOR_CONFIG_FILE
        chown root:root $APPARMOR_CONFIG_FILE
        aa-complain $DST_DIR/bin/pg
    else
        echo "It appears AppArmor is installed but apparmor-utils is not."
        echo "To enable the AppArmor profile, install apparmor-utils"
        echo "and run aa-complain $DST_DIR/bin/pg"
    fi
fi

echo
echo "Packet Guardian is now installed"
echo "Please edit the configuration to your"
echo "liking and them run 'service pg start'"
echo
