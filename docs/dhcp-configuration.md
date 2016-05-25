# DHCP Configuration

The DHCP configuration file syntax is a custom syntax loosely based the DHCPD format. The sample DHCP configuration includes explanations and examples of the possible formats. The integrated DHCP is customized for this specific application. It adheres to the base DHCP RFC, but not all DHCP options are implemented. The available options are:

- subnet-mask
- router
- domain-name-server
- domain-name
- broadcast-address
- network-time-protocol-servers

Options which allow for multiple values such as domain-name-server and network-time-protocol-servers, must be a list of comma separated values. E.g: `option domain-name-server 10.1.0.1, 10.1.0.2`.

To specify multiple ranges in a single subnet, each range must be in its own pool. This is contrary to DHCPD where a single pool can contain multiple ranges.
