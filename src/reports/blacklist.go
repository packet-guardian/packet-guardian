package reports

import (
	"errors"
	"net/http"
	"sort"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

func init() {
	RegisterReport("blackisted-users", "Blacklisted Users", blacklistedUsersReport)
	RegisterReport("blackisted-devices", "Blacklisted Devices", blacklistedDevicesReport)
}

func blacklistedUsersReport(e *common.Environment, w http.ResponseWriter, r *http.Request) error {
	sql := `SELECT "value" FROM "blacklist";`
	blkUserRows, err := e.DB.Query(sql)
	if err != nil {
		e.Log.WithField("ErrMsg", err).Error("SQL statement failed")
		return errors.New("SQL Query Failed")
	}
	defer blkUserRows.Close()

	blacklistedUsers := make([]*models.User, 0)

	for blkUserRows.Next() {
		var username string
		if err := blkUserRows.Scan(&username); err != nil {
			e.Log.WithField("ErrMsg", err).Error("")
			continue
		}
		user, err := models.GetUserByUsername(e, username)
		if err != nil {
			e.Log.WithField("ErrMsg", err).Error("")
			continue
		}
		blacklistedUsers = append(blacklistedUsers, user)
	}

	sort.Sort(models.UsernameSorter(blacklistedUsers))

	data := map[string]interface{}{
		"users": blacklistedUsers,
	}

	e.Views.NewView("report-blacklisted-users", r).Render(w, data)
	models.ReleaseUsers(blacklistedUsers)
	return nil
}

func blacklistedDevicesReport(e *common.Environment, w http.ResponseWriter, r *http.Request) error {
	devices, err := models.SearchDevicesByField(e, "blacklisted", "1")
	if err != nil {
		e.Log.WithField("ErrMsg", err).Error("")
		return errors.New("SQL Query Failed")
	}

	sort.Sort(models.MACSorter(devices))

	data := map[string]interface{}{
		"devices": devices,
	}

	e.Views.NewView("report-blacklisted-devices", r).Render(w, data)
	return nil
}
