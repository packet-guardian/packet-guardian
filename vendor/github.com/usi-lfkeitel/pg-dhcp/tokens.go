// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"

	"github.com/onesimus-systems/dhcp4"
)

type token int

const (
	tkEnd token = iota
	tkGlobal
	tkNetwork
	tkSubnet
	tkPool
	tkRegistered
	tkUnregistered

	tkServerIdentifier
	tkRange
	tkFreeLeaseAfter

	beginSettings
	tkOption
	tkDefaultLeaseTime
	tkMaxLeaseTime
	endSettings
)

var tokens = [...][]byte{
	tkEnd:          []byte("end"),
	tkGlobal:       []byte("global"),
	tkNetwork:      []byte("network"),
	tkSubnet:       []byte("subnet"),
	tkPool:         []byte("pool"),
	tkRegistered:   []byte("registered"),
	tkUnregistered: []byte("unregistered"),

	tkOption:           []byte("option"),
	tkDefaultLeaseTime: []byte("default-lease-time"),
	tkMaxLeaseTime:     []byte("max-lease-time"),
	tkServerIdentifier: []byte("server-identifier"),
	tkRange:            []byte("range"),
	tkFreeLeaseAfter:   []byte("free-lease-after"),
}

func isSetting(b []byte) bool {
	for i := beginSettings; i < endSettings; i++ {
		if bytes.Equal(b, tokens[i]) {
			return true
		}
	}
	return false
}

// This list contains only the options we need
// It will change as needs change
var options = map[string]dhcp4.OptionCode{
	"subnet-mask":                   1,
	"router":                        3,
	"domain-name-server":            6,
	"domain-name":                   15,
	"broadcast-address":             28,
	"network-time-protocol-servers": 42,
}
