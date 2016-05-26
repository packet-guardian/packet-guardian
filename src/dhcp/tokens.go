// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"

	"github.com/onesimus-systems/dhcp4"
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
	TkFreeLeaseAfter

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
	TkFreeLeaseAfter:   []byte("free-lease-after"),
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
