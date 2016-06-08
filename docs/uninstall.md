# Uninstall

## Uninstall Script Usage

```
uninstall.sh [flags]

    -p Purge all configuration files along with the main application files. THE DATABASE WILL BE DELETED.
    -y Answer "yes" to all confirmation prompts
```

## Uninstalling from install.sh

If you installed Packet Guardian using the install script, follow these steps to uninstall:

1. cd into the packet-guardian folder
2. run `scripts/uninstall.sh` as root

```Bash
$ cd /opt/packet-guardian
$ ./scripts/uninstall.sh
```

**NOTE**: This will only remove the program files. All configuration files, logs, and the database will remain untouched. AppArmor/SELinux profiles and Service files will also be removed.

## Purging configuration files

To remove everything dealing with Packet Guardian, use the flag `-p` when running uninstall.sh. This will delete **ALL** files originally created by Packet Guardian including the database file. If you want to keep the database, make a copy before running the uninstall script.

```Bash
$ ./scripts/uninstall.sh -p
```
