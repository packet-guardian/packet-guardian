# Using Docker

The packet-guardian Docker image contains the binaries for both the web management and DHCP server components. Which one is ran depends on how the container is started. This build of Packet Guardian DOES NOT support SQLite. MySQL is currently the only supported database when using the Docker image. Also, make sure it's running in ANSI mode.

A configuration file is not provided by default. One must be provided and available at `/app/packet-guardian/config/config.toml`. This path can be overridden with the environment variable `CONFIG_FILE`.

Here's an example of how to run the containers:

```shell
# Web Server
docker run \
    --name pg-web \
    -d \
    -p 80:80 \
    -p 443:443 \
    -v $PWD/config.toml:/app/packet-guardian/config/config.toml:ro \
    packet-guardian web

# DHCP Server
# The DHCP requires a separate configuration for it lease pools
docker run \
    --name pg-dhcp \
    -d \
    -p 67:67 \
    -v $PWD/config.toml:/app/packet-guardian/config/config.toml:ro \
    -v $PWD/dhcp.conf:/app/packet-guardian/config/dhcp.conf:ro \
    packet-guardian dhcp
```

Running MariaDB in ANSI mode:

```shell
docker run \
    --name mariadb \
    -d \
    -e MYSQL_ROOT_PASSWORD="password" \
    -e MYSQL_DATABASE="pg" \
    mariadb --ansi
```
