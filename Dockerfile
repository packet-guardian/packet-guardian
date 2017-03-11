FROM alpine:3.5

MAINTAINER Lee Keitel <lee@keitel.xyz>

ENV CONFIG_FILE config/config.toml

ADD dist/dist.tar.gz /app
ADD docker-entrypoint.sh /docker-entrypoint.sh

EXPOSE 67 80 443

WORKDIR /app/packet-guardian

ENTRYPOINT ["/docker-entrypoint.sh"]
