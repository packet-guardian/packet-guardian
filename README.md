Packet Guardian
---------------

[![Build Status](https://travis-ci.org/onesimus-systems/packet-guardian.svg?branch=master)](https://travis-ci.org/onesimus-systems/packet-guardian)

Packet Guardian is an easy to setup and use captive portal for wired or wireless networks.
It works in conjunction with a local DNS server and integrated DHCP server to redirect
clients to a registration page.

Setup
-----

TL;DR -

1. Get a copy of the distribution package from the Github releases page.
2. Untar it to a directory. The simplest is /opt.
3. Copy `config/config.sample.toml` to `config/config.toml`
4. Copy `config/dhcp-config.sample.conf` to `config/dhcp.conf`
5. Edit the configurations as needed. See the related documentation for each format.
6. Setup the database
7. Start Packet Guardian

TL;DR Code -

```bash
$ cd /opt
$ wget [URI to github download]
$ tar -xzf pg-[version].tar.gz
$ cd packet-guardian/config
$ cp config.sample.toml config.toml
$ vim config.toml
$ cp dhcp-config.sample.conf dhcp.conf
$ vim dhcp.conf
$ sqlite3 database.sqlite3 < db-schema-sqlite.sql
$ cd ..
$ bin/pg -c config/config.toml
```

Building From Source
--------------------

The distribution package is compiled against 64-bit Linux. However, the application
is written in Go with no OS specific system calls. This means you can download the
source and should be able to compile against any OS Go supports. Packet Guardian
has not been tested on other platforms. However I'm always willing to help if a
problem comes up.

To build from source, download the repo, cd into it and run `make dist`. The compiled
package will be in a tar file in the `dist` directory.

You can also run `make install` to install into the $GOPATH. Running `make build`
will build the binary and put it in the `bin` directory.

Building a Docker Image
-----------------------

This repo comes with a handy Dockerfile. To create a new image which will compile
PG from source, run `docker build -t pg --rm .`. It exposes ports 67 (DHCP), 80, and 443.
Port 443 is optional if you don't want to use HTTPS. You can run the container with:

```
docker run -it \
    --name guardian \
    -p 67:67 \
    -p 80:80 \
    -p 443:443 \
    pg
```

This will run Packet Guardian using a configuration located at
`/go/src/app/config/config.toml`. You can mount a volume to use a
custom configuration or edit the sample configuration before building the image.

Note, the sample configuration does not start the DHCP server. You will need to
create a custom configuration with DHCP enabled and create a DHCP configuration File
at `/go/src/app/config/dhcp.conf`.

Starting Packet Guardian
------------------------

The configuration file may be explicitly given to PG at runtime or it will look
for the file in a predefined set of directories. The order being:

- The `-c` flag when run from the command line
- The `PG_CONFIG` environment variable
- `./config.toml`
- `$HOME/.pg/config.toml`
- `/etc/packet-guardian/config.toml`

If a configuration file isn't found, PG will exit with an error.

Configuration File Syntax
-------------------------

The configuration file is written in [TOML](https://github.com/toml-lang/toml).
All available options are given in the sample configuration along with their defaults
and a short explanation.

DHCP Configuration Syntax
-------------------------

The DHCP configuration file syntax is a custom syntax loosely based the DHCPD format.
The sample DHCP configuration includes explanations and examples of the possible
formats. The integrated DHCP is customized for this specific application.
It adheres to the base DHCP RFC, but not all DHCP options are implemented. The
available options are:

- subnet-mask
- router
- domain-name-server
- domain-name
- broadcast-address
- network-time-protocol-servers

Options which allow for multiple values such as domain-name-server and network-time-protocol-servers,
must be a list of comma separated values. E.g: `option domain-name-server 10.1.0.1, 10.1.0.2`.

To specify multiple ranges in a single subnet, each range must be in its own pool.
This is contrary to DHCPD where a single pool can contain multiple ranges.

Using AppArmor
--------------

It's recommended to use AppArmor with Packet Guardian. AppArmor is a permissions
enforcement system on Ubuntu. It's equivalent to SELinux in CentOS and Red Hat.
An AppArmor profile has been included in the `config` directory. Here's how to
get it set up (run as root):

```bash
$ cd /opt/packet-guardian
$ cp config/apparmor.conf /etc/apparmor.d/opt.packet-guardian.bin.pg
$ chown root:root /etc/apparmor.d/opt.packet-guardian.bin.pg
$ aa-enforce /opt/packet-guardian/bin/pg
```

If you want to make sure everything is working correctly, you can use `aa-complain`
instead of `aa-enforce` which will allow the application to do anything but complain
if it's not following the profile. AppArmor messages will be shown in the syslog.

Installing as a Service
-----------------------

Using Packet Guardian as a service makes it a lot easier to manage. In the `config`
directory there is an upstart configuration file that can be used on Upstart
capable systems. Here's how to install it (run all as sudo):

1. Create a new user named `packetg`: `adduser -M packetg`
2. Change ownership of application files: `chown -R packetg:packetg /opt/packet-guardian`
3. Install service: `cd /opt/packet-guardian && cp config/upstart.conf /etc/init/pg.conf && chown root:root /etc/init/pg.conf`
4. Allow the application to bind to ports: `setcap 'cap_net_bind_service=+ep' /opt/packet-guardian/bin/pg`
5. Run application: `service pg start`
