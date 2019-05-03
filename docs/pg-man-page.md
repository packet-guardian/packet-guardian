# pg

## NAME

**pg** - web management frontend for Packet Guardian

## SYNOPSIS

**pg** [OPTIONS]

## DESCRIPTION

**pg** starts a web server to host the mangement frontend for Packet Guardian.
It's where clients are redirected to register a device and where users can login
to manage their own devices if the configuration allows.

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

**-v**, **-version**

    Print version information and exit.

## EXAMPLES

Starting a basic server:

```shell
pg -c config.toml
```

Checking the syntax of a configuration file:

```shell
pg -c config.toml -t
```
