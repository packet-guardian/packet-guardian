package dhcp

import (
	"database/sql"
	"errors"
	"net"
	"regexp"
	"strings"
	"time"
)

const macRegex = "^(?:[a-fA-F0-9]{2}\\:){6}$"

var (
	errAlreadyRegistered = errors.New("Device is already registered")
	errBlacklisted       = errors.New("Device or username is blacklisted")
	errMalformedMAC      = errors.New("Invalid MAC address")
)

// IsRegistered checks if a MAC address is registed in the database
func IsRegistered(db *sql.DB, mac string) bool {
	stmt, err := db.Prepare("SELECT \"id\" FROM \"device\" WHERE \"mac\" = ?")
	if err != nil {
		return false
	}
	var id int
	err = stmt.QueryRow(mac).Scan(&id)
	if err == nil {
		return true
	}
	return false
}

// GetMacFromIP finds the mac address that has the lease ip
func GetMacFromIP(ip net.IP, leasesFile string) (string, error) {
	l, err := getLeaseFromFile(ip, leasesFile)
	if err != nil {
		return "", err
	}
	return l.mac, nil
}

// Register a new device to a user
func Register(db *sql.DB, mac, user, platform, ip, ua, subnet string) error {
	mac = strings.ToLower(mac)
	if !isValidMac(mac) {
		return errMalformedMAC
	}
	if IsRegistered(db, mac) {
		return errAlreadyRegistered
	}
	regTime := time.Now().Unix()
	expires := 0
	stmt, err := db.Prepare("INSERT INTO \"device\" VALUES (null, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(mac, user, ip, platform, subnet, expires, regTime, ua)
	return err
}

func isValidMac(mac string) bool {
	matched, err := regexp.MatchString(macRegex, mac+":")
	if err != nil {
		return false
	}
	return matched
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
