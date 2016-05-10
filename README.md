Packet Guardian
---------------

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
6. Start Packet Guardian

TL;DR Code -

```shell
# cd /opt
# wget [URI to github download]
# tar -xzf pg-[version].tar.gz
# cd packet-guardian/config
# cp config.sample.toml config.toml
# vim config.toml
# cp dhcp-config.sample.conf dhcp.conf
# vim dhcp.conf
# cd ..
# bin/pg -c config/config.toml
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
