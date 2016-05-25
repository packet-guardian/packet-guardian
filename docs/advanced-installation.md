# Advanced Installation

## AppArmor

It's recommended to use AppArmor with Packet Guardian. AppArmor is a permissions enforcement system on Ubuntu. It's equivalent to SELinux in CentOS and Red Hat. An base AppArmor profile has been included in the `config` directory. You may need to edit it depending on how you setup Packet Guardian. The file `apparmor.conf` assumes PG is installed in `/opt/packet-guardian`. The file `apparmor-ext.conf` assumes you've run the script `$REPO_DIR/install.sh` which spreads out the configuration and run files into their appropriate places in the file system. Use which ever is appropiate for your situation.

Here's how to get start (run as root):

```bash
$ cd $PG_ROOT # By default /opt/packet-guardian
# Make sure to change the last part to match the real location of the binary
$ cp config/apparmor.conf /etc/apparmor.d/opt.packet-guardian.bin.pg
$ chown root:root /etc/apparmor.d/opt.packet-guardian.bin.pg
$ aa-complain /opt/packet-guardian/bin/pg
```

Using `aa-complain` is the safest to begin with. This will allow Packet Guardian to do anything it wants (within user and group permissions), but will warn in the syslog if it steps outside of its AppArmor profile. Adjust the profile as needed. Use the command `aa-enforce` to completely enable the profile.

## Running as a Service

If you used the script `$REPO_DIR/install.sh` this was done for you.

Using Packet Guardian as a service makes it a lot easier to manage. In the `config` directory there is an upstart configuration file that can be used on Upstart capable systems. Here's how to install it (run all as sudo):

1. Create a new user named `packetg`: `adduser -M packetg` The `-M` means don't create a home directory.
2. Change ownership of application files: `chown -R packetg:packetg /opt/packet-guardian`
3. Install service: `cd /opt/packet-guardian && cp config/upstart.conf /etc/init/pg.conf && chown root:root /etc/init/pg.conf`
4. Allow the application to bind to restricted ports: `setcap 'cap_net_bind_service=+ep' /opt/packet-guardian/bin/pg`
5. Run application: `service pg start`
