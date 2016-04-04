package main

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/onesimus-systems/packet-guardian/common"
	"github.com/onesimus-systems/packet-guardian/dhcp"
)

type device struct {
	MAC            string
	UA             string
	Platform       string
	IP             string
	DateRegistered string
}

func rootHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		reg, err := dhcp.IsRegisteredByIP(e.DB, net.ParseIP(ip), e.Config.DHCP.LeasesFile)
		if err != nil {
			e.Log.Errorf("Error checking auto registration IP: %s", err.Error())
		}
		if reg {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/register", http.StatusTemporaryRedirect)
		}
	}
}

func manageHandler(e *common.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := e.Sessions.GetSession(r, e.Config.Webserver.SessionName)
		username := sess.GetString("username")

		sql := "SELECT \"mac\", \"userAgent\", \"platform\", \"regIP\", \"dateRegistered\""
		sql += "FROM \"device\" WHERE \"username\" = ? ORDER BY \"dateRegistered\" ASC"
		rows, err := e.DB.Query(sql, username)
		if err != nil {
			e.Log.Error(err.Error())
		}

		devices := make([]device, 0)
		for rows.Next() {
			var mac string
			var ua string
			var platform string
			var regIP string
			var dateRegistered int64
			err := rows.Scan(&mac, &ua, &platform, &regIP, &dateRegistered)
			if err != nil {
				e.Log.Error(err.Error())
				continue
			}

			d := device{
				MAC:            mac,
				UA:             ua,
				Platform:       platform,
				IP:             regIP,
				DateRegistered: time.Unix(dateRegistered, 0).Format("01/02/2006 15:04:05"),
			}
			devices = append(devices, d)
		}

		data := struct {
			SiteTitle   string
			CompanyName string
			Devices     []device
		}{
			SiteTitle:   e.Config.Core.SiteTitle,
			CompanyName: e.Config.Core.SiteCompanyName,
			Devices:     devices,
		}
		e.Templates.ExecuteTemplate(w, "manage", data)
	}
}
