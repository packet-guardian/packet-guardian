package reports

import (
	"errors"
	"net"
	"net/http"
	"sort"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func init() {
	RegisterReport("blackisted-users", "Blacklisted Users", blacklistedUsersReport)
	RegisterReport("blackisted-devices", "Blacklisted Devices", blacklistedDevicesReport)
}

func blacklistedUsersReport(e *common.Environment, w http.ResponseWriter, r *http.Request, stores stores.StoreCollection) error {
	sql := `SELECT "value" FROM "blacklist";`
	blkUserRows, err := e.DB.Query(sql)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:blacklist",
		}).Error("SQL statement failed")
		return errors.New("SQL Query Failed")
	}
	defer blkUserRows.Close()

	var blacklistedUsers []*models.User

	for blkUserRows.Next() {
		var username string
		if err := blkUserRows.Scan(&username); err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "reports:blacklist",
			}).Error("Error scanning from SQL")
			continue
		}
		_, err := net.ParseMAC(username)
		if err == nil { // Probably a MAC address
			continue
		}
		user, err := stores.Users.GetUserByUsername(username)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":    err,
				"package":  "reports:blacklist",
				"username": username,
			}).Error("Error getting user")
			continue
		}
		blacklistedUsers = append(blacklistedUsers, user)
	}

	sort.Sort(models.UsernameSorter(blacklistedUsers))

	data := map[string]interface{}{
		"users": blacklistedUsers,
	}

	e.Views.NewView("report-blacklisted-users", r).Render(w, data)
	return nil
}

func blacklistedDevicesReport(e *common.Environment, w http.ResponseWriter, r *http.Request, stores stores.StoreCollection) error {
	sql := `SELECT "value" FROM "blacklist";`
	blkDevRows, err := e.DB.Query(sql)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:blacklist",
		}).Error("SQL statement failed")
		return errors.New("SQL Query Failed")
	}
	defer blkDevRows.Close()

	var devices []*models.Device

	for blkDevRows.Next() {
		var macAddr string
		if err := blkDevRows.Scan(&macAddr); err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "reports:blacklist",
			}).Error("Error scanning from SQL")
			continue
		}
		mac, err := net.ParseMAC(macAddr)
		if err != nil { // Not a mac address, probably a username
			continue
		}
		device, err := stores.Devices.GetDeviceByMAC(mac)
		if err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "reports:blacklist",
				"MAC":     macAddr,
			}).Error("Error getting user")
			continue
		}
		devices = append(devices, device)
	}

	sort.Sort(models.MACSorter(devices))

	data := map[string]interface{}{
		"devices": devices,
	}

	e.Views.NewView("report-blacklisted-devices", r).Render(w, data)
	return nil
}
