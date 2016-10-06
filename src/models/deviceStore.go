// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"net"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/pg-dhcp"
)

type DHCPDeviceStore struct {
	e *common.Environment
}

func NewDHCPDeviceStore(e *common.Environment) *DHCPDeviceStore {
	return &DHCPDeviceStore{e: e}
}

func (d *DHCPDeviceStore) GetDeviceByMAC(mac net.HardwareAddr) (dhcp.Device, error) {
	return GetDeviceByMAC(d.e, mac)
}
