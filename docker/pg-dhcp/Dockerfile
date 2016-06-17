FROM pg-base:latest

MAINTAINER Lee Keitel <lee@keitel.xyz>

EXPOSE 67

WORKDIR /app/packet-guardian

ENTRYPOINT ["/app/packet-guardian/bin/dhcp"]
CMD ["-c", "config/config.toml"]
