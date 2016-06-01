# Building

The distribution package is compiled against 64-bit Linux. However, the application is written in Go with no OS specific system calls however it does use cgo. This means you can download the source and should be able to compile against any OS Go supports so long has you have a recent gcc compiler installed. Packet Guardian has not been tested on other platforms. However I'm always willing to help if a problem comes up.

There are three different ways to build Packet Guardian:

- `make`
- `make dist`
- `make local-install`

All the instructions below assume you've cloned the repo and have `cd`ed into it.

## make

This command will run `go build` and output the executable to `$REPO_DIR/bin`. Although similar to `make local-install`, like `go install` it does not save compiled packages. Typically using `make local-install` is the preferred method since it will save compiled packages and use those if they haven't been edited saving valuable dev time.

## make local-install

This command will run `go install` but with $GOBIN set to `$REPO_DIR/bin`. Meaning you can then run Packet Guardian using `bin/pg` from the repo directory.

## make dist

This command will create a distributable tar file. The version number is determined by the latest git tag and commit. The new tar file will be in the dist folder. It can be extracted using `tar -xzf $path_to_file`. This will extract the archive into the folder `packet-guardian` in the current directory. For example, if you're in the folder `/opt` and run the tar command, Packet Guardian will be located at `/etc/packet-guardian`.
