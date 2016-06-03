# DHCP Configuration

## Overview

The DHCP configuration file syntax is a custom syntax loosely based the DHCPD format. The sample DHCP configuration includes explanations and examples of the possible formats. The DHCP server is customized for this specific application. It adheres to the base DHCP RFC, but not all DHCP options are implemented. Implemented options are described below.

Options which allow for multiple values such as domain-name-server and network-time-protocol-servers, must be a list of comma separated values. E.g: `option domain-name-server 10.1.0.1, 10.1.0.2`.

To specify multiple scopes in a single subnet, each scope must be in its own pool. This is contrary to DHCPD where a single pool can contain multiple range statements. See the Subnet and Pool sections for more.

## Global

Sample:

```
global
    option domain-name example.com
    server-identifier 10.0.0.1

    registered
        free-lease-after 172800
        default-lease-time 86400
        max-lease-time 86400
        option domain-name-server 10.1.0.1, 10.1.0.2
    end

    unregistered
        free-lease-after 600
        default-lease-time 360
        max-lease-time 360
        option domain-name-server 10.0.0.1
    end
end
```

The global section contains three subsections. The "root" which is any statement before a `registered` or `unregistered` keyword. Options specified here will be applied to every subnet regardless of registration status unless overridden elsewhere. Global root specific statements are:

- `server-identifier` - The IP address of the DHCP server

Registered and unregistered blocks may be located inside a global or network block. Settings here will apply to either registered or unregistered leases either globally (if in global block) or for any containing subnets. These blocks may contain any option along with the lease time settings.

## Options

Most options correspond to a DHCP option and begin with the keyword `option`. The available options are:

- `subnet-mask`
- `router`
- `domain-name-server` - May take a comma separate list of IP addresses
- `domain-name`
- `broadcast-address`
- `network-time-protocol-servers` - May take a comma separate list of IP addresses

The following options do NOT begin with the `option` keyword:

- `default-lease-time` - The amount of time in seconds a lease will be active for. Defaults to 12 hours.
- `max-lease-time` - The maximum amount of time in seconds a lease will be active for. Defaults to 12 hours.
- `free-lease-after` - The time in seconds that a lease will be paired with a client MAC address. If a client requests an address after this time, it is not guaranteed they will be given the same lease. This option can only be specified inside a registered/unregistered block within the global block.

## Network

```
network Network1
    unregistered
        ...
    end
    registered
        ...
    end
end
```

A network block groups multiple subnets into logical units. Although technically all subnets could be located in a single network block, it would be incredibly inefficient and difficult to determine true network usage.

The start line syntax is `network [name]`. The name is completely arbitrary but must be unique to each network block. Options may be specified within a network block in which case they will apply to both registered and unregistered leases in that network. Every network must contain a registered and unregistered block. And each block must have at lease one subnet.

## Subnet

```
# Shortened pool syntax - One pool/range
subnet 10.0.1.0/24
    range 10.0.1.10 10.0.1.200
    option router 10.0.1.1
end

# Full pool syntax - Multiple pools/ranges
subnet 10.0.2.0/24
    option router 10.0.2.1
    pool
        range 10.0.2.10 10.0.2.100
    end
    pool
        range 10.0.2.150 10.0.2.200
    end
end

# Invalid shortened pool syntax - range must be first
subnet 10.0.1.0/24
    option router 10.0.1.1
    range 10.0.1.10 10.0.1.200
end

# Invalid full pool syntax - range cannot appear outside of a pool block
subnet 10.0.2.0/24
    range 10.0.2.10 10.0.2.30
    option router 10.0.2.1
    pool
        range 10.0.2.40 10.0.2.100
    end
    pool
        range 10.0.2.150 10.0.2.200
    end
end
```

A subnet block forms the fundamental building block for the server. Each subnet must be with a registered or unregistered block inside a network block. The start line syntax is `subnet [ip range in CIDR notation]`. A subnet may contain any valid options as described above. A subnet may contain a single pool in which case a single range statement may be given. If using this syntax, the range statement but be directly after the subnet start line. Otherwise you will need to use the full pool syntax described below.

## Pool

A pool splits a subnet into multiple ranges. Typically, the shortened syntax will suffice where only one pool is present in a subnet. If you want multiple address ranges, you will need multiple pool blocks. Pool blocks may contain any valid option as specified above. Each pool must contain one range statement with the syntax `range [start address] [end address]`. The range is inclusive.
