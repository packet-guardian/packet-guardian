#! /usr/bin/env bash

# Check running as root
if [[ $UID -ne 0 ]]; then
    echo "This file must be ran as root."
    exit 1
fi

SRC_TAR="$1"
APP_DIR="/opt/packet-guardian"
UPSTART_SERVICE_DIR="/etc/init"
SYSTEMD_SERVICE_DIR="/etc/systemd/system"
LOG_DIR="/var/log/packet-guardian"
DATA_DIR="/var/lib/packet-guardian"
CONFIG_DIR="/etc/packet-guardian"
APPARMOR_DIR="/etc/apparmor.d"

SYSTEMD=""
APPARMOR_INSTALLED=""
APPARMOR_UTILS_INSTALLED=""

if which systemctl >/dev/null 2>&1; then
    SYSTEMD="t"
fi
if which apparmor_status >/dev/null 2>&1; then
    APPARMOR_INSTALLED="t"
fi
if which aa-complain >/dev/null 2>&1; then
    APPARMOR_UTILS_INSTALLED="t"
fi

stopService() {
    echo "Stopping any running instances"
    if [[ -n $SYSTEMD ]]; then
        systemctl stop pg-dhcp >/dev/null 2>&1
        systemctl stop pg >/dev/null 2>&1
    else
        service pg-dhcp stop >/dev/null 2>&1
        service pg stop >/dev/null 2>&1
    fi
}

confirm() {
    echo -n "$1 [y/N]: "
    read -n 1 imsure
    echo
    if [[ $imsure != "y" ]]; then
        exit 0
    fi
}

installed() {
    test -f $DATA_DIR/.installed
    return $?
}

installService() {
    cd $APP_DIR
    if [[ -n $SYSTEMD ]]; then
        echo "Installing Systemd Service"
        cp config/service/systemd/pg.service $SYSTEMD_SERVICE_DIR/pg.service
        cp config/service/systemd/dhcp.service $SYSTEMD_SERVICE_DIR/pg-dhcp.service
        chown root:root $SYSTEMD_SERVICE_DIR/pg.service
        chown root:root $SYSTEMD_SERVICE_DIR/pg-dhcp.service
        systemctl daemon-reload
        systemctl enable pg.service
        systemctl enable pg-dhcp.service
    else
        echo "Installing Upstart Service"
        cp config/service/upstart/pg.conf $UPSTART_SERVICE_DIR/pg.conf
        cp config/service/upstart/dhcp.conf $UPSTART_SERVICE_DIR/pg-dhcp.conf
        chown root:root $UPSTART_SERVICE_DIR/pg.conf
        chown root:root $UPSTART_SERVICE_DIR/pg-dhcp.conf
    fi
}

setKernalPermissions() {
    echo "Setting kernel permissions"
    setcap 'cap_net_bind_service=+ep' $APP_DIR/bin/pg
    setcap 'cap_net_bind_service=+ep' $APP_DIR/bin/dhcp
}

installAppArmorProfile() {
    # Install apparmor profile if available
    if [[ -n $APPARMOR_INSTALLED ]]; then
        echo "Installing AppArmor profile"
        mkdir -p $APPARMOR_DIR
        cp config/apparmor/pg/apparmor-ext.conf $APPARMOR_DIR/opt.packet-guardian.bin.pg
        cp config/apparmor/dhcp/apparmor-ext.conf $APPARMOR_DIR/opt.packet-guardian.bin.dhcp
        chown root:root $APPARMOR_DIR/opt.packet-guardian.bin.pg
        chown root:root $APPARMOR_DIR/opt.packet-guardian.bin.dhcp
        if [[ -n $APPARMOR_UTILS_INSTALLED ]]; then
            aa-complain $APP_DIR/bin/pg
            aa-complain $APP_DIR/bin/dhcp
        else
            echo "It appears AppArmor is installed but apparmor-utils is not."
            echo "To enable the AppArmor profile, install apparmor-utils"
            echo "and run:"
            echo "aa-complain $APP_DIR/bin/pg"
            echo "aa-complain $APP_DIR/bin/dhcp"
        fi
    else
        echo "AppArmor doesn't appear to be installed. Skipping."
    fi
}

setPermissions() {
    chown -R packetg:packetg $APP_DIR
    chown -R packetg:packetg $LOG_DIR
    chown -R packetg:packetg $DATA_DIR
    chown -R root:packetg $CONFIG_DIR
}

upgrade() {
    if ! installed; then
        echo "It appears Packet Guardian is not installed."
        echo "Please install Packet Guardian before trying to upgrade."
        echo
        exit 1
    fi

    cd /opt

    if [[ ! -d $APP_DIR ]]; then
        echo "It appears Packet Guardian is not installed."
        echo "Please install Packet Guardian before trying to upgrade."
        echo
        exit 1
    else
        echo "Moving current installation to $APP_DIR.old"
        rm -rf $APP_DIR.old
        mv $APP_DIR $APP_DIR.old
    fi

    SRC_TAR="$(realpath $SRC_TAR)"
    stopService

    echo "Extracting new version"
    tar -xzf $SRC_TAR
    chown -R packetg:packetg $APP_DIR

    echo "Copying configuration files"
    cp $APP_DIR/config/config-dhcp.sample.toml $CONFIG_DIR
    cp $APP_DIR/config/config-pg.sample.toml $CONFIG_DIR
    cp $APP_DIR/config/config-dhcp.sample.toml $CONFIG_DIR/config-dhcp.toml.dist
    cp $APP_DIR/config/config-pg.sample.toml $CONFIG_DIR/config-pg.toml.dist
    cp $APP_DIR/config/dhcp-config.sample.toml $CONFIG_DIR
    cp $APP_DIR/config/policy.txt $CONFIG_DIR/policy.txt.dist

    # Perform any necessary SQL migrations
    # sqlite3 $DATA_DIR/database.sqlite3 < $APP_DIR/config/db-schema-sqlite.sql

    setPermissions
    installService
    setKernalPermissions
    installAppArmorProfile

    echo
    echo "Packet Guardian is now upgraded"
    echo "Please check the docs and configuration"
    echo "for new options and release notes."
    echo "When you're ready, start Packet Guardian using:"
    echo
    echo "service pg start OR systemctl start pg"
    echo "service pg-dhcp start OR systemctl start pg-dhcp"
    echo
}

install() {
    if [[ ! -d $APP_DIR ]]; then
        echo "It appears Packet Guardian is not in the correct place."
        echo "Please extract the Packet Guardian release to $APP_DIR"
        echo "and try again."
        echo
        exit 1
    fi

    if installed; then
        echo "It appears Packet Guardian is already installed."
        confirm "This will overwrite all configuration files. Are you sure?"
    fi

    echo "Creating packetg user"
    id -u packetg >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        useradd -M packetg
    fi

    echo "Creating data directories"
    mkdir -p $LOG_DIR
    mkdir -p $DATA_DIR
    mkdir -p $CONFIG_DIR
    echo "Creating configuration files"
    cp $APP_DIR/config/config-dhcp.sample.toml $CONFIG_DIR
    cp $APP_DIR/config/config-pg.sample.toml $CONFIG_DIR
    cp $APP_DIR/config/config-dhcp.sample.toml $CONFIG_DIR/config-dhcp.toml
    cp $APP_DIR/config/config-pg.sample.toml $CONFIG_DIR/config-pg.toml
    cp $APP_DIR/config/dhcp-config* $CONFIG_DIR
    cp $APP_DIR/config/policy.txt $CONFIG_DIR

    sqlite3 $DATA_DIR/database.sqlite3 < $APP_DIR/config/db-schema-sqlite.sql

    setPermissions
    installService
    setKernalPermissions
    installAppArmorProfile

    touch $DATA_DIR/.installed

    echo
    echo "Packet Guardian is now installed"
    echo "Please edit the configurations to your"
    echo "liking and them run using:"
    echo
    echo "service pg start OR systemctl start pg"
    echo "service pg-dhcp start OR systemctl start pg-dhcp"
    echo
}

if [[ -z $(which sqlite3) ]]; then
    echo "sqlite3 is required to install Packet Guardian"
    exit 1
fi

if [[ -z $SRC_TAR ]]; then
    confirm "This will install Packet Guardian. Are you sure?"
    install
else
    confirm "This will upgrade Packet Guardian. Are you sure?"
    if [[ ! -f $SRC_TAR ]]; then
        echo "The source tarball doesn't appear to exist, please check your path and try again"
        exit 1
    fi
    upgrade
fi
