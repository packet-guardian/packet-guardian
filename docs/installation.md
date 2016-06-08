# Installation

## Automatic Installation (recommended)

1. Get a copy of the distribution package from the Github releases page.
2. Untar it to /opt.
3. Run `scripts/install.sh`
4. Edit the configurations as needed. See the related documentation for each format.
5. Start Packet Guardian

```bash
# Run as root
$ cd /opt
$ wget https://github.com/onesimus-systems/packet-guardian/releases/latest/pg-dist-$VERSION.tar.gz
$ tar -xzf pg-dist-$VERSION.tar.gz
$ cd packet-guardian
$ ./scripts/install.sh
```

The install script will place the configuration files in `/etc/packet-guardian`, the SQLite database at `/var/lib/packet-guardian/database.sqlite3`, and the log files at `/var/log/packet-guardian`. Edit the files in `/etc/packet-guardian` to suit your environment. The file config-pg.toml will be read by the webserver component and config-dhcp.toml will be read by the DHCP server. config-dhcp.toml is basically a stripped down version of the full config. You can use the same configuration for both binaries, but you will need to edit the Upstart/Systemd service files to do so.

## Manually

Before going to much further, it is highly recommended that you use the automatic installation script. It will ensure everything is set up correctly and in its proper place. Only use the manual method if you absolutely have to.

1. Get a copy of the distribution package from the Github releases page.
2. Extract the tarball to a directory. It will create a packet-guardian folder when extracted.
3. The configuration files are located in the config directory. There are sample files for the webserver component and the DHCP component. It's recommended to copy the sample files and edit as needed. The .toml files are the given to the binary at runtime. The dhcp-config.conf file is where the DHCP scopes and server settings are defined. See the relevant documentation for those file formats. You WILL NEED to edit the configuration and set the correct paths. The defaults paths assume the installation method is the automatic install script.
4. Initialize the database but running `sqlite3 config/database.sqlite3 < config/db-schema-sqlite.sql`.
4. To start the system, run the following commands with the current directory being the packet-guardian folder extracted earlier: `bin/pg -c $path_to_config` and `bin/dhcp -c $path_to_config`. Both binaries take the .toml file as the -c option. The path to the DHCP scope configuration is defined in the toml file.
5. To use AppArmor on Ubuntu or other AppArmor systems, see the [Additional Installation](additional-installation.md) documentation.
6. To setup the binaries as services, see the [Additional Installation](additional-installation.md) documentation.

## Configuration file location

The configuration file may be explicitly given to PG at runtime or it will look for the file in a predefined set of directories. The order being:

- The `-c` flag when run from the command line
- The `PG_CONFIG` environment variable - This is how the service files give the configuration file
- `$PWD/config.toml`
- `$HOME/.pg/config.toml`
- `/etc/packet-guardian/config.toml`

If a configuration file isn't found, PG will exit with an error.
