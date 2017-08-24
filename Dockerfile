# Build binaries
FROM golang:1.8 AS build

ENV PROJ_ROOT=/go/src/github.com/packet-guardian/packet-guardian
ENV DIST_FILENAME=dist.tar.gz
# CGO must be disabled to work properly in Alpine
ENV CGO_ENABLED=0

COPY . $PROJ_ROOT

RUN cd $PROJ_ROOT \
    && go get -u github.com/jteeuwen/go-bindata/... \
    && make dist \
    && cd / \
    && tar -xzf $PROJ_ROOT/dist/dist.tar.gz

# Build final application image
FROM alpine:3.6

ARG version
ARG builddate
ARG vcsref

LABEL name="lfkeitel/packet-guardian" \
      version="$version" \
      build-date="$builddate" \
      vcs-type="git" \
      vcs-url="https://github.com/packet-guardian/packet-guardian" \
      vcs-ref="$vcsref" \
      maintainer="Lee Keitel <lfkeitel@usi.edu>"

ENV CONFIG_FILE config/config.toml

COPY --from=build /packet-guardian /app/packet-guardian
ADD docker-entrypoint.sh /docker-entrypoint.sh

EXPOSE 67 80 443

WORKDIR /app/packet-guardian

ENTRYPOINT ["/docker-entrypoint.sh"]
