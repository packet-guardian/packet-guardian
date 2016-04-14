package dhcp

import (
	"database/sql"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/onesimus-systems/packet-guardian/common"
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
		baseSQL += " OR \"value\"=?"
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

// AddToBlacklist add values to the blacklist
func AddToBlacklist(db *sql.DB, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	baseSQL := "INSERT INTO \"blacklist\" (\"value\") VALUES"
	for range values {
		baseSQL += " (?),"
	}
	baseSQL = strings.TrimRight(baseSQL, ",")
	stmt, err := db.Prepare(baseSQL)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(values...)
	return err
}

// RemoveFromBlacklist removes values from the blacklist
func RemoveFromBlacklist(db *sql.DB, values ...interface{}) error {
	if len(values) == 0 {
		return nil
	}

	baseSQL := "DELETE FROM \"blacklist\" WHERE 1=0"
	for range values {
		baseSQL += " OR \"value\" = ?"
	}
	stmt, err := db.Prepare(baseSQL)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(values...)
	return err
}

// GetBlacklist will retreive values from the blacklist if they exist.
// Calling with no filter will return the fill blacklist.
func GetBlacklist(db *sql.DB, filter ...interface{}) ([]string, error) {
	baseSQL := "SELECT \"value\" FROM \"blacklist\""
	var rows *sql.Rows
	var err error

	if len(filter) > 0 {
		baseSQL += " WHERE 1=0"
		for range filter {
			baseSQL += " OR \"value\"=?"
		}
		rows, err = db.Query(baseSQL, filter...)
	} else {
		rows, err = db.Query(baseSQL)
	}

	if err != nil {
		return []string{}, err
	}

	var results []string
	for rows.Next() {
		var val string
		err := rows.Scan(&val)
		if err != nil {
			continue
		}
		results = append(results, val)
	}
	return results, nil
}

func getDeviceByID(db *sql.DB, ids ...interface{}) ([]Device, error) {
	sql := "SELECT \"id\", \"mac\", \"userAgent\", \"platform\", \"regIP\", \"dateRegistered\", \"username\" FROM \"device\" WHERE 0=1"

	for range ids {
		sql += " OR \"id\" = ?"
	}

	rows, err := db.Query(sql, ids...)
	if err != nil {
		return nil, err
	}

	bl, err := GetBlacklist(db)
	if err != nil {
		return nil, err
	}

	var results []Device
	for rows.Next() {
		var id int
		var mac string
		var ua string
		var platform string
		var regIP string
		var dateRegistered int64
		var username string
		err := rows.Scan(&id, &mac, &ua, &platform, &regIP, &dateRegistered, &username)
		if err != nil {
			return nil, err
		}

		r := Device{
			ID:             id,
			MAC:            mac,
			UserAgent:      ua,
			Platform:       platform,
			RegIP:          regIP,
			DateRegistered: time.Unix(dateRegistered, 0).Format("01/02/2006 15:04:05"),
			Username:       username,
			Blacklisted:    common.StringInSlice(mac, bl),
		}
		results = append(results, r)
	}
	return results, nil
}
