FROM pg-base:latest

MAINTAINER Lee Keitel <lee@keitel.xyz>

EXPOSE 80 443

WORKDIR /app/packet-guardian

ENTRYPOINT ["/app/packet-guardian/bin/pg"]
CMD ["-c", "config/config.toml"]
