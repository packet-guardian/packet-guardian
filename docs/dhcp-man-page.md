# dhcp

## NAME

**dhcp** - DHCP server for Packet Guardian

## SYNOPSIS

**dhcp** [OPTIONS]

## DESCRIPTION

**dhcp** starts and runs the DHCP server for Packet Guardian. It's primary usage is to
separate "registered" from "unregistered" clients and return leases depending on a
device's state.

## OPTIONS

**-c** path

    Set the configuration file. If this option is not specified, pg will search for a
    configuration file using the following order:
        - `PG_CONFIG` environment variable
        - `$PWD/config.toml`
        - `$PWD/config/config.toml`
        - `$HOME/.pg/config.toml`
        - `/etc/packet-guardian/config.toml`

**-d**

    Run in debug/development mode.

**-t**

    Test the configuration for syntax errors and exit.

**-td**

    Test the DHCP pool configuration for syntax errors. Requires the **-c** option.

**-v**, **-version**

    Print version information and exit.

## EXAMPLES

Starting a basic server:

```shell
dhcp -c config.toml
```

Checking the syntax of a configuration file:

```shell
dhcp -c config.toml -t
```

Checking the syntax of a pool configuration file:

```shell
dhcp -c config.toml -td
```
