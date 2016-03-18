package dhcp

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/onesimus-systems/net-guardian/auth"
	"github.com/onesimus-systems/net-guardian/common"
)

const macRegex = "(?:[a-f0-9]{2}\\:){6}"

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
	baseSQL := "SELECT \"id\" FROM \"blacklist\" WHERE 1=0"
	for range values {
		baseSQL = baseSQL + " OR \"value\" = ?"
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

// RegisterHTTPHandler serves and handles the registration page for an end user
func RegisterHTTPHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.Templates.ExecuteTemplate(w, "register.tmpl", nil)
	}
}

// AutoRegisterHandler handles the path /register/auto
func AutoRegisterHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		if !auth.IsValidLogin(e.DB, username, r.FormValue("password")) {
			e.Log.Errorf("Failed authentication for %s", username)
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"Incorrect username or password"})
			return
		}

		ip := strings.Split(r.RemoteAddr, ":")[0]
		mac, err := GetMacFromIP(net.ParseIP(ip), e.Config.DHCP.LeasesFile)
		if err != nil {
			e.Log.Errorf("Failed to get MAC for IP %s: %s", ip, err.Error())
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"Error detected MAC address"})
			return
		}

		bl, err := IsBlacklisted(e.DB, mac, username)
		if err != nil {
			e.Log.Errorf("There was an error checking the blacklist for MAC %s and user %s", mac, username)
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"There was an error registering your device."})
		}
		if bl {
			e.Log.Errorf("Attempted authentication of blacklisted MAC or user %s - %s", mac, username)
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"There was an error registering your device. BL"})
			return
		}

		err = Register(e.DB, mac, username, r.FormValue("platform"), ip, r.UserAgent(), "")
		if err != nil {
			e.Log.Errorf("Failed to register MAC address %s to user %s: %s", mac, username, err.Error())
			e.Templates.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{"There was an error registering your device"})
			return
		}
		e.Log.Infof("Successfully registered MAC %s to user %s", mac, username)
		e.Templates.ExecuteTemplate(w, "success.tmpl", struct{ Message string }{"Device successfully registered"})
	}
}
