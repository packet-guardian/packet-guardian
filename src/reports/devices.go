package reports

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func init() {
	RegisterReport("device-last-seen", "Device Last Seen", deviceLastSeenReport)
}

func deviceLastSeenReport(e *common.Environment, w http.ResponseWriter, r *http.Request, stores stores.StoreCollection) error {
	since, exists := r.URL.Query()["since"]
	if !exists {
		since = []string{time.Now().Format(time.DateOnly)}
	}

	sinceTime, err := time.Parse(time.DateOnly, since[0])
	if err != nil {
		sinceTime = time.Now().Add(-200 * time.Hour)
	}

	pageNum := 1
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		pageNum = page
	}

	lastseen := sinceTime.Unix()

	op, has := r.URL.Query()["op"]
	if has && op[0] == "download-report" {
		return downloadDeviceLastSeenReport(w, stores.Devices, lastseen)
	}

	devices, err := stores.Devices.Search(
		`last_seen < ? ORDER BY "last_seen" ASC LIMIT ?,?`,
		lastseen, (common.PageSize*pageNum)-common.PageSize, common.PageSize,
	)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:devices",
		}).Error("Failed to get devices")
		return nil
	}

	resultCnt, err := deviceLastSeenResultCnt(e, lastseen)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:leases",
		}).Error("Failed to get devices")
		return nil
	}

	pageEnd := pageNum * common.PageSize
	if resultCnt < pageEnd {
		pageEnd = resultCnt
	}

	data := map[string]interface{}{
		"devices": devices,
		"since":   sinceTime.Format(time.DateOnly),

		"resultCnt":   resultCnt,
		"page":        pageNum,
		"hasNextPage": pageNum*common.PageSize < resultCnt,
		"pageStart":   ((pageNum - 1) * common.PageSize) + 1,
		"pageEnd":     pageEnd,
	}

	e.Views.NewView("admin-report-device-last-seen", r).Render(w, data)
	return nil
}

func deviceLastSeenResultCnt(e *common.Environment, lastseen int64) (int, error) {
	sql := `SELECT count(*) as "device_count" FROM "device" WHERE last_seen < ?`
	row := e.DB.QueryRow(sql, lastseen)
	var deviceLastSeenResultCnt int
	err := row.Scan(&deviceLastSeenResultCnt)
	if err != nil {
		return 0, err
	}
	return deviceLastSeenResultCnt, nil
}

func downloadDeviceLastSeenReport(w http.ResponseWriter, dstore stores.DeviceStore, lastseen int64) error {
	devices, err := dstore.Search(
		`last_seen < ? ORDER BY "last_seen" ASC`,
		lastseen,
	)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)
	csvWriter.Write([]string{"id", "mac", "description", "last-seen", "registered"})

	for _, d := range devices {
		csvWriter.Write([]string{
			strconv.Itoa(d.ID),
			d.MAC.String(),
			d.Description,
			d.LastSeen.Format(time.RFC3339),
			d.DateRegistered.Format(time.RFC3339),
		})
	}
	csvWriter.Flush()
	return csvWriter.Error()
}
