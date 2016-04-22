package dhcp

import (
	"bytes"

	"github.com/krolaw/dhcp4"
)

type Token int

const (
	TkEnd Token = iota
	TkGlobal
	TkNetwork
	TkSubnet
	TkPool
	TkRegistered
	TkUnregistered

	TkServerIdentifier
	TkRange

	beginSettings
	TkOption
	TkDefaultLeaseTime
	TkMaxLeaseTime
	endSettings
)

var tokens = [...][]byte{
	TkEnd:          []byte("end"),
	TkGlobal:       []byte("global"),
	TkNetwork:      []byte("network"),
	TkSubnet:       []byte("subnet"),
	TkPool:         []byte("pool"),
	TkRegistered:   []byte("registered"),
	TkUnregistered: []byte("unregistered"),

	TkOption:           []byte("option"),
	TkDefaultLeaseTime: []byte("default-lease-time"),
	TkMaxLeaseTime:     []byte("max-lease-time"),
	TkServerIdentifier: []byte("server-identifier"),
	TkRange:            []byte("range"),
}

func isSetting(b []byte) bool {
	for i := beginSettings; i < endSettings; i++ {
		if bytes.Equal(b, tokens[i]) {
			return true
		}
	}
	return false
}

// This list is honestly just thrown together
// It will change as needed
var options = map[string]dhcp4.OptionCode{
	"subnet-mask":                   1,
	"time-offset":                   2,
	"router":                        3,
	"domain-name-server":            6,
	"log-server":                    7,
	"hostname":                      12,
	"boot-file-size":                13,
	"domain-name":                   15,
	"swap-server":                   16,
	"root-path":                     17,
	"extensions-path":               18,
	"interface-mtu":                 26,
	"all-subnets-are-Local ":        27,
	"broadcast-address":             28,
	"perform-mask-discovery":        29,
	"mask-supplier":                 30,
	"perform-router-discovery":      31,
	"router-solicitation-address":   32,
	"static-route":                  33,
	"network-time-protocol-servers": 42,
	"relay-agent-information":       82,
	"ipaddress-lease-time":          51,
	"renewal-time-value":            58,
	"rebinding-time-value":          59,
	"vendor-class-identifier":       60,
	"client-identifier":             61,
	"tftp-server-name":              66,
	"boot-file-name":                67,
}
