package reports

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lfkeitel/verbose/v4"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
	"github.com/packet-guardian/packet-guardian/src/stats"
)

func init() {
	RegisterReport("lease-stats", "Lease Statistics", leaseReport)
}

func leaseReport(e *common.Environment, w http.ResponseWriter, r *http.Request) error {
	network, ok := r.URL.Query()["network"]
	if !ok || len(network) != 1 {
		networks := stats.DHCPNetworkList(e)
		sort.Strings(networks)

		data := map[string]interface{}{
			"networks": networks,
		}

		e.Views.NewView("reports-leases-list", r).Render(w, data)
		return nil
	}

	networkName := network[0]
	_, registered := r.URL.Query()["registered"]

	pageNum := 1
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		pageNum = page
	}

	endTime := time.Now().Unix()
	leases, err := stores.GetLeaseStore(e).SearchLeases(
		`network = ? AND registered = ? AND end > ? ORDER BY "mac" ASC LIMIT ?,?`,
		networkName, registered, endTime, (common.PageSize*pageNum)-common.PageSize, common.PageSize,
	)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:leases",
		}).Error("Failed to get leases")
		return nil
	}

	leaseCnt, err := leaseCount(e, networkName, registered, endTime)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:leases",
		}).Error("Failed to get leases")
		return nil
	}

	pageEnd := pageNum * common.PageSize
	if leaseCnt < pageEnd {
		pageEnd = leaseCnt
	}

	sort.Sort(models.LeaseSorter(leases))

	data := map[string]interface{}{
		"network":    strings.Title(networkName),
		"registered": registered,
		"leases":     leases,

		"leaseCnt":    leaseCnt,
		"page":        pageNum,
		"hasNextPage": pageNum*common.PageSize < leaseCnt,
		"pageStart":   ((pageNum - 1) * common.PageSize) + 1,
		"pageEnd":     pageEnd,
	}

	e.Views.NewView("reports-leases", r).Render(w, data)
	return nil
}

func leaseCount(e *common.Environment, networkName string, registered bool, end int64) (int, error) {
	sql := `SELECT count(*) as "lease_count" FROM "lease" WHERE "network" = ? AND "registered" = ? AND "end" > ?`
	row := e.DB.QueryRow(sql, networkName, registered, end)
	var leaseCount int
	err := row.Scan(&leaseCount)
	if err != nil {
		return 0, err
	}
	return leaseCount, nil
}
