# Using Docker

Docker compatibility is currently experimental and most likely broken. Please use at your own risk. If anything goes wrong please file an issue but understand that Docker compatibility is not high priority at this time.

This repo comes with a handy Dockerfile. To create a new image which will compile PG from source, run `docker build -t pg --rm .`. It exposes ports 67 (DHCP), 80, and 443. Port 443 is optional if you don't want to use HTTPS. You can run the container with:

```
docker run -it \
    --name guardian \
    -p 67:67 \
    -p 80:80 \
    -p 443:443 \
    pg
```

This will run Packet Guardian using a configuration located at `/go/src/app/config/config.toml`. You can mount a volume to use a custom configuration or edit the sample configuration before building the image.

Note, the sample configuration does not start the DHCP server. You will need to create a custom configuration with DHCP enabled and create a DHCP configuration File at `/go/src/app/config/dhcp.conf`.
