// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"

type multiple int // How many of a token type are allowed
type length int   // The number of bytes an option can be in length

const (
	oneOrMore multiple = -1 // One or more of a token are allowed

	unlimited length = 0 // The option may be as long as needed (within spec)
)

type optionSchema struct {
	token      token
	multi      multiple // How many of the token are allowed
	multipleOf int
	maxlen     length // Maximum number of bytes the option can be
}

var (
	// The following are all the schemas used in the options defined below
	booleanSchema  = &optionSchema{token: BOOLEAN, multi: 1, maxlen: 1, multipleOf: 1}
	singleIPSchema = &optionSchema{token: IP_ADDRESS, multi: 1, maxlen: 4, multipleOf: 4}
	multiIPSchema  = &optionSchema{token: IP_ADDRESS, multi: oneOrMore, maxlen: unlimited, multipleOf: 4}
	stringSchema   = &optionSchema{token: STRING, multi: 1, maxlen: unlimited, multipleOf: 1}
	int8Schema     = &optionSchema{token: NUMBER, multi: 1, maxlen: 1, multipleOf: 1}
	int16Schema    = &optionSchema{token: NUMBER, multi: 1, maxlen: 2, multipleOf: 2}
	int32Schema    = &optionSchema{token: NUMBER, multi: 1, maxlen: 4, multipleOf: 4}
	anySchema      = &optionSchema{token: ANY, multi: oneOrMore, maxlen: unlimited, multipleOf: 1}
)

type option struct {
	dhcp4.Option
	vendor bool
}

type dhcpOptionBlock struct {
	code   dhcp4.OptionCode
	schema *optionSchema
}

var options = map[string]*dhcpOptionBlock{
	// Standard BOOTP options
	"subnet-mask": &dhcpOptionBlock{
		code:   dhcp4.OptionSubnetMask,
		schema: singleIPSchema,
	},
	"time-offset": &dhcpOptionBlock{
		code:   dhcp4.OptionTimeOffset,
		schema: int32Schema,
	},
	"router": &dhcpOptionBlock{
		code:   dhcp4.OptionRouter,
		schema: multiIPSchema,
	},
	"time-server": &dhcpOptionBlock{
		code:   dhcp4.OptionTimeServer,
		schema: multiIPSchema,
	},
	"name-server": &dhcpOptionBlock{
		code:   dhcp4.OptionNameServer,
		schema: multiIPSchema,
	},
	"domain-name-server": &dhcpOptionBlock{
		code:   dhcp4.OptionDomainNameServer,
		schema: multiIPSchema,
	},
	"log-server": &dhcpOptionBlock{
		code:   dhcp4.OptionLogServer,
		schema: multiIPSchema,
	},
	"cookie-server": &dhcpOptionBlock{
		code:   dhcp4.OptionCookieServer,
		schema: multiIPSchema,
	},
	"lpr-server": &dhcpOptionBlock{
		code:   dhcp4.OptionLPRServer,
		schema: multiIPSchema,
	},
	"impress-server": &dhcpOptionBlock{
		code:   dhcp4.OptionImpressServer,
		schema: multiIPSchema,
	},
	"resource-location-server": &dhcpOptionBlock{
		code:   dhcp4.OptionResourceLocationServer,
		schema: multiIPSchema,
	},
	"hostname": &dhcpOptionBlock{
		code:   dhcp4.OptionHostName,
		schema: stringSchema,
	},
	"boot-file-size": &dhcpOptionBlock{
		code:   dhcp4.OptionBootFileSize,
		schema: int16Schema,
	},
	"merit-dump-file": &dhcpOptionBlock{
		code:   dhcp4.OptionMeritDumpFile,
		schema: stringSchema,
	},
	"domain-name": &dhcpOptionBlock{
		code:   dhcp4.OptionDomainName,
		schema: stringSchema,
	},
	"swap-server": &dhcpOptionBlock{
		code:   dhcp4.OptionSwapServer,
		schema: singleIPSchema,
	},
	"root-path": &dhcpOptionBlock{
		code:   dhcp4.OptionRootPath,
		schema: stringSchema,
	},
	"extensions-path": &dhcpOptionBlock{
		code:   dhcp4.OptionExtensionsPath,
		schema: stringSchema,
	},

	// IP Layer Parameters per Host
	"ip-forwarding-toggle": &dhcpOptionBlock{
		code:   dhcp4.OptionIPForwardingEnableDisable,
		schema: booleanSchema,
	},
	"non-local-source-routing-toggle": &dhcpOptionBlock{
		code:   dhcp4.OptionNonLocalSourceRoutingEnableDisable,
		schema: booleanSchema,
	},
	"policy-filter": &dhcpOptionBlock{
		code:   dhcp4.OptionPolicyFilter,
		schema: multiIPSchema,
	},
	"max-datagram-reassembly-size": &dhcpOptionBlock{
		code:   dhcp4.OptionMaximumDatagramReassemblySize,
		schema: int16Schema,
	},
	"default-ip-ttl": &dhcpOptionBlock{
		code:   dhcp4.OptionDefaultIPTimeToLive,
		schema: int8Schema,
	},
	"path-mtu-aging-timeout": &dhcpOptionBlock{
		code:   dhcp4.OptionPathMTUAgingTimeout,
		schema: int32Schema,
	},
	"path-mtu-plateau-table": &dhcpOptionBlock{
		code:   dhcp4.OptionPathMTUPlateauTable,
		schema: &optionSchema{token: NUMBER, multi: oneOrMore, maxlen: unlimited},
	},

	// IP Layer Parameters per Interface
	"interface-mtu": &dhcpOptionBlock{
		code:   dhcp4.OptionInterfaceMTU,
		schema: int16Schema,
	},
	"all-subnets-are-local": &dhcpOptionBlock{
		code:   dhcp4.OptionAllSubnetsAreLocal,
		schema: booleanSchema,
	},
	"broadcast-address": &dhcpOptionBlock{
		code:   dhcp4.OptionBroadcastAddress,
		schema: singleIPSchema,
	},
	"perform-mask-discovery": &dhcpOptionBlock{
		code:   dhcp4.OptionPerformMaskDiscovery,
		schema: booleanSchema,
	},
	"mask-supplier": &dhcpOptionBlock{
		code:   dhcp4.OptionMaskSupplier,
		schema: booleanSchema,
	},
	"perform-router-discovery": &dhcpOptionBlock{
		code:   dhcp4.OptionPerformRouterDiscovery,
		schema: booleanSchema,
	},
	"router-solicitation-address": &dhcpOptionBlock{
		code:   dhcp4.OptionRouterSolicitationAddress,
		schema: singleIPSchema,
	},
	"static-route": &dhcpOptionBlock{
		code:   dhcp4.OptionStaticRoute,
		schema: multiIPSchema,
	},

	// Link Layer Parameters per Interface
	"trailer-encapsulation": &dhcpOptionBlock{
		code:   dhcp4.OptionTrailerEncapsulation,
		schema: booleanSchema,
	},
	"arp-cache-timeout": &dhcpOptionBlock{
		code:   dhcp4.OptionARPCacheTimeout,
		schema: int32Schema,
	},
	"ethernet-encapsulation": &dhcpOptionBlock{
		code:   dhcp4.OptionEthernetEncapsulation,
		schema: booleanSchema,
	},

	// TCP Parameters
	"tcp-default-ttl": &dhcpOptionBlock{
		code:   dhcp4.OptionTCPDefaultTTL,
		schema: int8Schema,
	},
	"tcp-keepalive-interval": &dhcpOptionBlock{
		code:   dhcp4.OptionTCPKeepaliveInterval,
		schema: int32Schema,
	},
	"tcp-keepalive-garbage": &dhcpOptionBlock{
		code:   dhcp4.OptionTCPKeepaliveGarbage,
		schema: booleanSchema,
	},

	// Application and Service Parameters
	"network-information-service-domain": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkInformationServiceDomain,
		schema: stringSchema,
	},
	"network-information-servers": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkInformationServers,
		schema: multiIPSchema,
	},
	"network-time-protocol-servers": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkTimeProtocolServers,
		schema: multiIPSchema,
	},
	"vendor-options": &dhcpOptionBlock{
		code:   dhcp4.OptionVendorSpecificInformation,
		schema: booleanSchema, // Enabled in config, generated for client
	},
	"netbios-over-tcpip-name-server": &dhcpOptionBlock{
		code:   dhcp4.OptionNetBIOSOverTCPIPNameServer,
		schema: multiIPSchema,
	},
	"netbios-over-tcpip-datagram-distribution-server": &dhcpOptionBlock{
		code:   dhcp4.OptionNetBIOSOverTCPIPDatagramDistributionServer,
		schema: multiIPSchema,
	},
	"netbios-over-tcpip-node-type": &dhcpOptionBlock{
		code:   dhcp4.OptionNetBIOSOverTCPIPNodeType,
		schema: int8Schema,
	},
	"netbios-over-tcpip-scope": &dhcpOptionBlock{
		code:   dhcp4.OptionNetBIOSOverTCPIPScope,
		schema: stringSchema,
	},
	"xwindow-system-font-server": &dhcpOptionBlock{
		code:   dhcp4.OptionXWindowSystemFontServer,
		schema: multiIPSchema,
	},
	"xwindow-system-display-manager": &dhcpOptionBlock{
		code:   dhcp4.OptionXWindowSystemDisplayManager,
		schema: multiIPSchema,
	},
	"nis+-Domain": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkInformationServicePlusDomain,
		schema: stringSchema,
	},
	"nis+-Servers": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkInformationServicePlusServers,
		schema: multiIPSchema,
	},
	"mobile-ip-home-agent": &dhcpOptionBlock{
		code:   dhcp4.OptionMobileIPHomeAgent,
		schema: multiIPSchema,
	},
	"simple-mail-transport-protocol": &dhcpOptionBlock{
		code:   dhcp4.OptionSimpleMailTransportProtocol,
		schema: multiIPSchema,
	},
	"post-office-protocol-server": &dhcpOptionBlock{
		code:   dhcp4.OptionPostOfficeProtocolServer,
		schema: multiIPSchema,
	},
	"network-news-transport-protocol": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkNewsTransportProtocol,
		schema: multiIPSchema,
	},
	"default-www-server": &dhcpOptionBlock{
		code:   dhcp4.OptionDefaultWorldWideWebServer,
		schema: multiIPSchema,
	},
	"default-finger-server": &dhcpOptionBlock{
		code:   dhcp4.OptionDefaultFingerServer,
		schema: multiIPSchema,
	},
	"default-irc-server": &dhcpOptionBlock{
		code:   dhcp4.OptionDefaultInternetRelayChatServer,
		schema: multiIPSchema,
	},
	"street-talk-server": &dhcpOptionBlock{
		code:   dhcp4.OptionStreetTalkServer,
		schema: multiIPSchema,
	},
	"street-talk-directory-assistance": &dhcpOptionBlock{
		code:   dhcp4.OptionStreetTalkDirectoryAssistance,
		schema: multiIPSchema,
	},

	// DHCP Extensions
	"tftp-server-name": &dhcpOptionBlock{
		code:   dhcp4.OptionTFTPServerName,
		schema: stringSchema,
	},
	"renewal-time-value": &dhcpOptionBlock{
		code:   dhcp4.OptionRenewalTimeValue,
		schema: int32Schema,
	},
	"rebinding-time-value": &dhcpOptionBlock{
		code:   dhcp4.OptionRebindingTimeValue,
		schema: int32Schema,
	},
}
