# Upgrading

## Installed Via Automatic Script

If Packet Guardian was installed using the install script, upgrading is as simple as one command. When Packet Guardian was installed, a command called `pg-upgrade` was also installed that helps with upgrading.

```Bash
$ wget https://github.com/packet-guardian/packet-guardian/releases/latest/pg-dist-$VERSION.tar.gz
$ pg-upgrade pg-dist-$VERSION.tar.gz
```

You can skip the confirmation prompt by using the `-y` flag before the tar filename.

## Manually Installed

To upgrade Packet Guardian manually, move the current installation to a different folder, untar the new version, and apply any necassary edits to configuration files if they weren't located outside of the main folder.

```Bash
$ wget https://github.com/packet-guardian/packet-guardian/releases/latest/pg-dist-$VERSION.tar.gz
$ cd /opt
$ mv packet-guardian packet-guardian.old
$ tar -xzf pg-dist-$VERSION.tar.gz
```
