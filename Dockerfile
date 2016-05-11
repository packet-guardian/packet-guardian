FROM golang:1.6-wheezy

MAINTAINER Lee Keitel <lee@keitel.xyz>

RUN apt-get update \
    && apt-get install -y --no-install-recommends -y sqlite3 \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/github.com/onesimus-systems/packet-guardian \
    && ln -s /go/src/github.com/onesimus-systems/packet-guardian /go/src/app
WORKDIR /go/src/github.com/onesimus-systems/packet-guardian

COPY . /go/src/app
RUN ["/bin/sh", "-c", "make install"]
RUN /usr/bin/sqlite3 config/database.sqlite3 < config/db-schema-sqlite.sql
RUN cp config/config.sample.toml config/config.toml

EXPOSE 67 80 443

ENTRYPOINT ["/go/bin/pg", "-c", "config/config.toml"]
