// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"net"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

type DeviceStore struct {
	e *common.Environment
}

func NewDeviceStore(e *common.Environment) *DeviceStore {
	return &DeviceStore{e: e}
}

func (d *DeviceStore) GetDeviceByMAC(mac net.HardwareAddr) (*Device, error) {
	return GetDeviceByMAC(d.e, mac)
}
