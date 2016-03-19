package dhcp

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const macRegex = "^(?:[a-fA-F0-9]{2}\\:){6}$"

var (
	errAlreadyRegistered = errors.New("Device is already registered")
	errBlacklisted       = errors.New("Device or username is blacklisted")
	errMalformedMAC      = errors.New("Invalid MAC address")
	errGenericRegister   = errors.New("Failed to register device")
	errNoUsernameGiven   = errors.New("No username given")
	hostMutex            = sync.Mutex{}
)

// IsRegistered checks if a MAC address is registed in the database
func IsRegistered(db *sql.DB, mac string) (bool, error) {
	stmt, err := db.Prepare("SELECT \"id\" FROM \"device\" WHERE \"mac\" = ?")
	if err != nil {
		return false, err
	}
	var id int
	err = stmt.QueryRow(mac).Scan(&id)
	if err == nil {
		return true, nil
	}
	return false, nil
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
	if user == "" {
		return errNoUsernameGiven
	}
	mac = strings.ToLower(mac)
	if !isValidMac(mac) {
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

// StartHostWriteService spins off a goroutine and creates communication channels
// to write out a new DHCPd host file from the database db to the file filepath.
// The first returned channel is used to write a new file the second is to
// signal a quit. Both channels are buffered to a single value as the design
// is that no matter how many write requests are made while it's currently writting,
// it will only rewrite it once. This prevents a case where multiple requests are
// put in when it's not necassary.
func StartHostWriteService(db *sql.DB, filepath string) (chan bool, chan bool) {
	c := make(chan bool, 1)
	e := make(chan bool, 1)
	go hostWriteService(db, filepath, c, e)
	return c, e
}

func hostWriteService(db *sql.DB, filepath string, c <-chan bool, quit <-chan bool) {
	for {
		select {
		case <-quit:
			return
		case <-c:
			writeHostFile(db, filepath)
		}
	}
}

func writeHostFile(db *sql.DB, filepath string) error {
	hostMutex.Lock()
	defer hostMutex.Unlock()

	expired := time.Now().Unix()
	sql := "SELECT \"mac\", \"username\", \"userAgent\", \"regIP\", \"registered\" FROM \"device\" WHERE \"expired\" = 0 OR \"expired\" > ? ORDER BY \"username\" ASC"
	rows, err := db.Query(sql, expired)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filepath+"~", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close() // In case something goes wrong

	i := 1
	currentUser := ""
	for rows.Next() {
		var mac string
		var user string
		var ua string
		var regIP string
		var registered int64
		err := rows.Scan(&mac, &user, &ua, &regIP, &registered)
		if err != nil {
			return err
		}
		if user != currentUser {
			i = 1
			currentUser = user
		}
		regTime := time.Unix(registered, 0).Format(time.RFC1123)
		line := fmt.Sprintf("host %s-%d { hardware ethernet %s; }#%s#%s#%s\n", user, i, mac, ua, regTime, regIP)
		file.WriteString(line)
		i++
	}

	// Close the hosts file and move temp file in place
	file.Close()
	os.Remove(filepath)
	if err := os.Rename(filepath+"~", filepath); err != nil {
		return err
	}

	return nil
}
