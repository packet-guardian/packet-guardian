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

MAINTAINER Lee Keitel <lee@keitel.xyz>

ENV CONFIG_FILE config/config.toml

COPY --from=build /packet-guardian /app/packet-guardian
ADD docker-entrypoint.sh /docker-entrypoint.sh

EXPOSE 67 80 443

WORKDIR /app/packet-guardian

ENTRYPOINT ["/docker-entrypoint.sh"]
