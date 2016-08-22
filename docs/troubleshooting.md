# Troubleshooting Packet Guardian Issues

These troubleshooting tips were created with the assumption that Packet Guardian was installed using the provided install script. Please make any corrections to path or service names as needed for non-standard installations.

## DHCP Component

### DHCP Server won't start

- If you're trying to start it as a service, run the binary directly and see what error it gives.
- Make sure the configuration file is in the correct place and is valid.
- Make sure the binary has permission to bind to port 67. Since it's a lower "protected" port, only root users or binaries with special permission may access it.
- If using AppArmor, try disabling the profile. If it starts successfully, put the profile into complain mode and check syslog for what is causing a profile violation.

### DHCP server not giving out leases

- Make sure the service is running: `service pg-dhcp status` or `systemctl status pg-dhcp`.
- Check that UDP port 67 is open on the firewall for incoming traffic.
- Check that the relay is configured to the correct IP address for Packet Guardian.
- Check the subnet configuration and make sure all the IP ranges are correct.
- Check the DHCP logs to see if leases are being offered but not arriving at the client. (/var/log/pg_dhcp_log.log by default)

## Web Component

### Web Server won't start

- If you're trying to start it as a service, run the binary directly and see what error it gives.
- Make sure the configuration file is in the correct place and is valid.
- If using Nginx as a proxy, make sure the port is not in the protected range. (I.e. higher than 1024)
- If using AppArmor, try disabling the profile. If it starts successfully, put the profile into complain mode and check syslog for what is causing a profile violation.

### Web page isn't loading

- Make sure Nginx is running: `service nginx status`.
- Make sure the Packet Guardian web component is running: `service pg status`.
- Check the logs for any errors. (/var/log/pg_log.log by default)
- Check Nginx's logs for errors. (/var/log/nginx/error.log)
- Check Nginx's configuration and make sure the proxy port is correct.
- Check Packet Guardian's configuration to make sure it's using the correct proxy port.

## Other Stack Items

### Nginx won't start

- Check Nginx's logs for errors.
- Make sure the TLS certificate and key files exist (if applicable)
- Check the configuration syntax.

### BindDNS won't start

- Check Bind's logs to start errors.
- Make sure the zone file is in the correct location.
- Check the zone file syntax.

### User's aren't being redirected

- Make sure Bind is running.
- Make sure the "fake" Bind server is the only DNS server set for unregistered subnets.
- Make sure the client machine doesn't have manually set DNS servers.
- Make sure Bind is configured to resolve all domain names to the IP address of Packet Guardian.
- Make sure the Nginx configuration is set to redirect all domains to the true domain name.
