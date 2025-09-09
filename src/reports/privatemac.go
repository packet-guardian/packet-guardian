package reports

import (
	"errors"
	"net/http"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func init() {
	RegisterReport("private-devices", "Private MAC Addresses", privateMACReport)
}

type privateMACRow struct {
	Username, MAC string
}

func privateMACReport(e *common.Environment, w http.ResponseWriter, r *http.Request, stores stores.StoreCollection) error {
	sql := `SELECT "mac","username" FROM "device" WHERE "mac" REGEXP '^[0-9a-f][26ae]' ORDER BY "mac" ASC;`
	blkDevRows, err := e.DB.Query(sql)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:privatemac",
		}).Error("SQL statement failed")
		return errors.New("SQL Query Failed")
	}
	defer blkDevRows.Close()

	var devices []privateMACRow

	for blkDevRows.Next() {
		var macAddr, username string
		if err := blkDevRows.Scan(&macAddr, &username); err != nil {
			e.Log.WithFields(verbose.Fields{
				"error":   err,
				"package": "reports:privatemac",
			}).Error("Error scanning from SQL")
			continue
		}
		devices = append(devices, privateMACRow{
			Username: username,
			MAC:      macAddr,
		})
	}

	data := map[string]any{
		"devices": devices,
	}

	e.Views.NewView("admin-report-private-devices", r).Render(w, data)
	return nil
}
