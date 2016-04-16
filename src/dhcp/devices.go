package dhcp

import (
	"net"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

// IsRegistered checks if a MAC address is registed in the database.
// IsRegistered will return false if an error occurs as well as the error itself.
func IsRegistered(e *common.Environment, mac net.HardwareAddr) (bool, error) {
	device, err := models.GetDeviceByMAC(e, mac)
	if err != nil {
		return false, err
	}
	return (device.ID != 0), nil
}

// IsRegisteredByIP checks if an IP is leased to a registered MAC address.
// IsRegisteredByIP will return false if an error occurs as well as the error itself.
func IsRegisteredByIP(e *common.Environment, ip net.IP) (bool, error) {
	mac, err := GetMacFromIP(ip)
	if err != nil {
		return false, err
	}
	return IsRegistered(e, mac)
}

// GetMacFromIP finds the mac address that has the lease ip
func GetMacFromIP(ip net.IP) (net.HardwareAddr, error) {
	// TODO: Replace with getting lease info from database
	// l, err := getLeaseFromFile(ip, leasesFile)
	// if err != nil {
	// 	return net.HardwareAddr{}, err
	// }
	//return l.mac, nil
	return net.HardwareAddr{}, nil
}
