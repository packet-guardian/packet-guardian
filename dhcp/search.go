package dhcp

import (
	"net"
	"time"

	"github.com/onesimus-systems/packet-guardian/common"
)

// Query represents a search query for a mac, ip, or username. Only one field
// can be searched at a time. If multiple are given, the precident is MAC, IP, User
type Query struct {
	IP   net.IP
	MAC  net.HardwareAddr
	User string
}

func (q Query) Search(e *common.Environment) []Result {
	sql := "SELECT \"mac\", \"userAgent\", \"platform\", \"regIP\", \"dateRegistered\", \"username\" FROM \"device\" "
	param := ""

	if q.MAC != nil {
		sql += "WHERE \"mac\" = ?"
		param = q.MAC.String()
	} else if q.IP != nil {
		// Search for a lease
		// TODO: Finish when the lease system is added
		e.Log.Error("IP lease search not supported")
		return nil
	} else if q.User != "" {
		sql += "WHERE \"username\" LIKE ?"
		param = q.User + "%"
	} else {
		sql += "WHERE 1 = 1 OR \"username\" = ?"
	}

	sql += " ORDER BY \"username\" ASC"

	rows, err := e.DB.Query(sql, param)
	if err != nil {
		e.Log.Error(err.Error())
	}

	var results []Result
	for rows.Next() {
		var mac string
		var ua string
		var platform string
		var regIP string
		var dateRegistered int64
		var username string
		err := rows.Scan(&mac, &ua, &platform, &regIP, &dateRegistered, &username)
		if err != nil {
			e.Log.Error(err.Error())
			continue
		}

		r := Result{
			MAC:            mac,
			UserAgent:      ua,
			Platform:       platform,
			RegIP:          regIP,
			DateRegistered: time.Unix(dateRegistered, 0).Format("01/02/2006 15:04:05"),
			Username:       username,
		}
		results = append(results, r)
	}
	return results
}

type Result struct {
	MAC            string
	Platform       string
	RegIP          string
	DateRegistered string
	Username       string
	UserAgent      string
}
