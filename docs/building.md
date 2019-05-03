# Building

The distribution package is compiled against 64-bit Linux. However, the
application should build for any Go runtime target.

## Requirements

- Go 1.8+
- Yarn

## TL;DR

```shell
# Build the static assets
yarn install
yarn run build:prod

# Build the application binary
make
```

## Customizing the build

The build can be customized by which database backends are compiled. This is
controlled using the following build tags:

- `dball` (default): Compile all database backends.
- `dbmysql`: Compile the MySQL database backend.

Example building only MySQL: `BUILDTAGS=dbmysql make`. Or without CGO:
`CGO_ENABLED=0 BUILDTAGS=dbmysql make`.

## Make Commands

All the instructions below assume you've cloned the repo and have `cd`ed into
it.

### make

Run the test suite and build both components.

### make management

Build only the web management interface.

### make doc

Start a local godoc server.

### make fmt

Run `go fmt`.

### make test

Run test suite.

### make coverage

Run coverage tests.

### make benchmark

Run benchmarks.

### make lint

Run `golint`. Requires `github.com/golang/lint` to be installed: `go get
github.com/golang/lint/golint`

### make vet

Run `go vet`.

### make dist

This command will create a distributable tar file. The version number is
determined by the latest git tag and commit. The new tar file will be in the
dist folder. It can be extracted using `tar -xzf $path_to_file`. This will
extract the archive into the folder `packet-guardian` in the current directory.
For example, if you're in the folder `/opt` and run the tar command, Packet
Guardian will be located at `/etc/packet-guardian`.

### make clean

Clean up built binaries, logs, session data, etc.

### make docker

Build the Packet Guardian Docker image.
