# Additional Installation

## Running as a Service

If you used the script `scripts/install.sh` this was done for you.

Run these commands as root.

### Upstart - Ubuntu < 15.04

```bash
$ useradd -M packetg
$ chown -R packetg:packetg /opt/packet-guardian
$ cp $APP_DIR/config/service/upstart/pg.conf /etc/init/pg.conf
$ cp $APP_DIR/config/service/upstart/dhcp.conf /etc/init/pg-dhcp.conf
$ chown root:root /etc/init/pg.conf
$ chown root:root /etc/init/pg-dhcp.conf
$ service pg start
$ service pg-dhcp start
```

### Systemd - Ubuntu >= 15.04, Fedora, CentOS, Arch

```bash
$ useradd -M packetg
$ chown -R packetg:packetg /opt/packet-guardian
$ cp $APP_DIR/config/service/systemd/pg.service /etc/systemd/system/pg.service
$ cp $APP_DIR/config/service/systemd/dhcp.service /etc/systemd/system/pg-dhcp.service
$ chown root:root /etc/systemd/system/pg.service
$ chown root:root /etc/systemd/system/pg-dhcp.service
$ systemctl daemon-reload
$ systemctl start pg
$ systemctl start pg-dhcp
$ systemctl enable pg.service
$ systemctl enable pg-dhcp.service
```

If Packet Guardian is not starting, you may need to run `setcap 'cap_net_bind_service=+ep' /opt/packet-guardian/bin/pg` and `setcap 'cap_net_bind_service=+ep' /opt/packet-guardian/bin/dhcp`. This will set kernal permissions to allow the binaries to bind to restricted ports.

## AppArmor

If you used the script `scripts/install.sh` this was done for you.

It's recommended to use AppArmor with Packet Guardian. AppArmor is a permissions enforcement system on Ubuntu. It's equivalent to SELinux in CentOS and Red Hat. A base AppArmor profile has been included in the `config` directory. You may need to edit it depending on how you setup Packet Guardian. There are separate profiles for both the main pg and dhcp binaries. The file `apparmor.conf` assumes PG is installed in `/opt/packet-guardian`. The file `apparmor-ext.conf` assumes you've run the script `$REPO_DIR/install.sh` which spreads out the configuration and puts files into their appropriate places in the file system. Use which ever is appropriate for your situation.

Here's how to get start (run as root):

```bash
$ cd $PG_ROOT # By default /opt/packet-guardian
# Make sure to change the last part to match the real location of the binary
$ cp config/apparmor.conf /etc/apparmor.d/opt.packet-guardian.bin.pg
$ chown root:root /etc/apparmor.d/opt.packet-guardian.bin.pg
$ aa-complain /opt/packet-guardian/bin/pg
```

Using `aa-complain` is the safest to begin with. This will allow Packet Guardian to do anything it wants (within user and group permissions), but will warn in the syslog if it steps outside of its AppArmor profile. Adjust the profile as needed. Use the command `aa-enforce` to completely enable the profile.
