# Using Docker

**Docker compatibility is currently experimental. Please use at your own risk. If anything goes wrong please file an issue but understand that Docker compatibility is not high priority at this time.**

This repo contains three docker files. A base and then one each for the webserver and dhcp. Although they are similar, per Docker philosophy they are designed to be run as separate containers.

The DHCP container exposes port 67. The web server exposes ports 80 and 443.

To use these containers you will need to specify a configuration file and a database file. They can be specified however you like either via host-based volumes or container volumes. So long as the paths are correct it doesn't matter. The application is located at `/app/packet-guardian` which is also set as its working directory. Remember to give the same database to both containers otherwise the two pieces can't communicate. Also, using this method does not spread the files like a normal install. Meaning the install script is not ran and you will need to setup the SQLite database file.

Here's an example of how to run the containers:

```Bash
# Setup database
sqlite3 database.sqlite3 < /path-to-base-scheme.sql

# Web Server
docker run \
    --name packetg-web \
    -p 80:80 \
    -p 443:443 \
    -v $PWD/config.toml:/app/packet-guardian/config/config.toml:ro \
    -v $PWD/database.sqlite3:/app/packet-guardian/config/database.sqlite3 \
    pg-web

# DHCP Server
docker run \
    --name packetg-web \
    -p 67:67 \
    -v $PWD/config.toml:/app/packet-guardian/config/config.toml:ro \
    -v $PWD/database.sqlite3:/app/packet-guardian/config/database.sqlite3 \
    pg-dhcp
```
