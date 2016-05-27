# Installation

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

The configuration file may be explicitly given to PG at runtime or it will look for the file in a predefined set of directories. The order being:

- The `-c` flag when run from the command line
- The `PG_CONFIG` environment variable
- `./config.toml`
- `$HOME/.pg/config.toml`
- `/etc/packet-guardian/config.toml`

If a configuration file isn't found, PG will exit with an error.
