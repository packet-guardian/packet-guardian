package dhcp

import (
	"database/sql"
	"errors"
	"net"
	"sync"
	"time"
)

var (
	errAlreadyRegistered = errors.New("Device is already registered")
	errBlacklisted       = errors.New("Device or username is blacklisted")
	errMalformedMAC      = errors.New("Invalid MAC address")
	errGenericRegister   = errors.New("Failed to register device")
	errNoUsernameGiven   = errors.New("No username given")
	hostMutex            = sync.Mutex{}
)

// IsRegistered checks if a MAC address is registed in the database.
// IsRegistered will return false if an error occurs as well as the error itself.
func IsRegistered(db *sql.DB, mac net.HardwareAddr) (bool, error) {
	stmt, err := db.Prepare("SELECT \"id\" FROM \"device\" WHERE \"mac\" = ?")
	if err != nil {
		return false, err
	}
	var id int
	err = stmt.QueryRow(mac.String()).Scan(&id)
	if err == nil {
		return true, nil
	}
	return false, nil
}

// IsRegisteredByIP checks if an IP is leased to a registered MAC address.
// IsRegisteredByIP will return false if an error occurs as well as the error itself.
func IsRegisteredByIP(db *sql.DB, ip net.IP) (bool, error) {
	mac, err := GetMacFromIP(ip)
	if err != nil {
		return false, err
	}
	return IsRegistered(db, mac)
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

// Register a new device to a user. This function will check if the MAC address is valid
// and if it is already registered. This function does not enforce the blacklist.
func Register(db *sql.DB, mac net.HardwareAddr, user, platform string, ip net.IP, ua, subnet string) error {
	if user == "" {
		return errNoUsernameGiven
	}
	if mac.String() == "" {
		return errMalformedMAC
	}
	r, err := IsRegistered(db, mac)
	if err != nil {
		return errGenericRegister
	}
	if r {
		return errAlreadyRegistered
	}
	regTime := time.Now().Unix()
	expires := 0
	stmt, err := db.Prepare("INSERT INTO \"device\" VALUES (null, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(mac.String(), user, ip.String(), platform, subnet, expires, regTime, ua)
	return err
}

// FormatMacAddress will attempt to format and parse a string as a MAC address
func FormatMacAddress(mac string) (net.HardwareAddr, error) {
	// If no punctuation was provided, use the format xxxx.xxxx.xxxx
	if len(mac) == 12 {
		mac = mac[0:4] + "." + mac[4:8] + "." + mac[8:12]
	}
	return net.ParseMAC(mac)
}

// IsBlacklisted checks if a username or MAC is in the blacklist
func IsBlacklisted(db *sql.DB, values ...interface{}) (bool, error) {
	if len(values) == 0 {
		return false, nil
	}

	baseSQL := "SELECT \"id\" FROM \"blacklist\" WHERE 1=0"
	for range values {
		baseSQL = baseSQL + " OR \"value\"=?"
	}
	stmt, err := db.Prepare(baseSQL)
	if err != nil {
		return false, err
	}
	var id int
	err = stmt.QueryRow(values...).Scan(&id)
	if err == nil {
		return true, nil
	}
	return false, nil
}
