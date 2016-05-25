# Building

The distribution package is compiled against 64-bit Linux. However, the application is written in Go with no OS specific system calls however it does use cgo. This means you can download the source and should be able to compile against any OS Go supports so long has you have a recent gcc compiler installed. Packet Guardian has not been tested on other platforms. However I'm always willing to help if a problem comes up.

There are three different ways to build Packet Guardian:

    - `make dist`
    - `make install`
    - `make build`

All the instructions below assume you've cloned the repo and have `cd`ed into it. All instructions will also run tools such as `go vet` and `go test`.

## make dist

This command will create a distributable tar file. To specify a version number you can add `VERSION=x.y.z` to the command: `make dist VERSION=0.6.0`. The new tar file will be in the dist folder. It can be extracted using `tar -xzf $path_to_file`. This will extract the archive into the folder `packet-guardian` in the current directory. For example, if you're in the folder `/opt` and run the tar command, Packet Guardian will be located at `/etc/packet-guardian`.

## make install

This command will run `go install` but with $GOBIN set to `$REPO_DIR/bin`. Meaning you can then run Packet Guardian using `bin/pg` from the repo directory.

## make build

This command will run `go build` and output the executable to `$REPO_DIR/bin`. Although similar to `make install`, like `go install` it does not save compiled packages. Typically using `make installed` is the preferred method since it will save compiled packages and use those if they haven't been edited.
