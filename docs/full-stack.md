# Full Captive Portal Stack

Packet Guardian is just one (or two) pieces of the puzzle. To get a fully working and integrated captive portal you'll need two other pieces of software. A DNS server, and a webserver. Bind and Nginx will be used throughout as examples. Any DNS and any webserver software should work fine so long as the webserver can wildcard match domain names.

## Setting up the Network

Packet Guardian requires at least one vlan. Local subnets are currently not supported. All DHCP must be done via relays. The vlan needs to have two IP subnets assigned to it. The primary subnet will be for registered clients. The secondary is unregistered. A DHCP relay must be setup on the router/wireless controller to forward DHCP packets to Packet Guardian.

If the server is going to be in a restricted network segment (recommended), the following ports will need to be allowed in: tcp/udp 53 (DNS), 67 (DHCP server), 80, and 443 (web). The server will need to be able to unicast back to relay agents.

## Setting up Packet Guardian

The following configuration settings need to be changed:

- Use port 8080 (or any unprivileged port) for HTTP
- Disable HTTPS by commenting out the TLS settings
- Disable redirectHttpToHttps by commenting it or setting it to false
- Set the siteDomainName to the correct domain

Restart the Packet Guardian service.

## Setting up Bind

Bind needs to be setup such that every request to it is resolved to the address of the Packet Guardian server. Here's an example zone file:

```
$TTL 0
. IN SOA localhost.  root.localhost. (
        201605236       ; serial
        10800           ; refresh
        3600            ; retry
        604800          ; expire
        86400           ; default_ttl
)
                  IN      NS      wifi.example.com.

wifi.example.com.   IN      A       10.0.0.1
*.com.          IN      CNAME   wifi.example.com.
*.              IN      CNAME   wifi.example.com.
```

Change `wifi.example.com` to the domain name of the server. Change `10.0.0.1` to the real IP address of the server. Make sure to keep `$TTL` set to 0. This should prevent clients from caching the response. If you don't use a .com, make sure to change the first wildcard `*.com.` to the TLD of the real domain. For example, if you were using `wifi.example.edu`, it would be changed to `*.edu.`.

In named.conf, make sure to have the following:

```
zone "." {
    type master;
    file "zones/named.ca";
};
```

Change the file path to match the zone file if needed. It's relative to the directory option.

## Setting up Nginx

Nginx needs to be setup to redirect all domains to the domain of the registration server. There will be three server blocks. A wildcard, the http version of the correct domain, and the https version of the correct domain. HTTPS is not required but highly recommended. A certificate can be issued by [Let's Encrypt](https://letsencrypt.org). Here's the site configuration:

```Nginx
# Redirect wifi.example.com to HTTPS and preserve the request_uri
server {
        listen 80;
        server_name wifi.example.com;
        return 301 https://wifi.example.com$request_uri;
}

# Redirect all other domains to the main page
server {
        listen 80 default_server;
        server_name _; # wildcard on all domains
        # Instruct client not to cache redirect
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma no-cache;
        add_header Expires 0;
        return 307 https://wifi.example.com;
}

# Handle TLS for wifi.example.com
server {
        listen 443;
        server_name wifi.example.com;

        ssl on;
        ssl_certificate /path/to/tls.cert;
        ssl_certificate_key /path/to/tls.key;

        ssl_session_timeout 5m;

        # These settings are from https://cipherli.st/
        ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
        ssl_prefer_server_ciphers on;
        ssl_ciphers "EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH";
        ssl_ecdh_curve secp384r1; # Requires nginx >= 1.1.0
        ssl_session_cache shared:SSL:10m;
        ssl_stapling on; # Requires nginx >= 1.3.7
        ssl_stapling_verify on; # Requires nginx => 1.3.7
        add_header Strict-Transport-Security "max-age=63072000; includeSubdomains; preload";
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;

        location / {
                # Proxy all requests to Packet Guardian running on a higher port
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header Host $host;
                proxy_pass http://127.0.0.1:8080;
        }
}
```

## Conclusions

There are quite a few moving parts but when everything is in place it makes for nearly seamless portal system. In testing we've found that most modern devices automatically detect the need for the user to authenticate and will prompt the user to login. This provides a smooth transition for the user and with a simple disconnect, reconnect are back up and running.
