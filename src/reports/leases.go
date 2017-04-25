package reports

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func init() {
	RegisterReport("lease-stats", "Lease Statistics", leaseReport)
}

func leaseReport(e *common.Environment, w http.ResponseWriter, r *http.Request) error {
	network, ok := r.URL.Query()["network"]
	if !ok || len(network) != 1 {
		return nil
	}
	networkName := network[0]
	_, registered := r.URL.Query()["registered"]

	leases, err := stores.GetLeaseStore(e).SearchLeases(
		"network = ? AND registered = ? AND end > ?",
		networkName, registered, time.Now().Unix(),
	)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:leases",
		}).Error("Failed to get leases")
		return nil
	}

	sort.Sort(models.LeaseSorter(leases))

	data := map[string]interface{}{
		"network":    strings.Title(networkName),
		"registered": registered,
		"leases":     leases,
	}

	e.Views.NewView("reports-leases", r).Render(w, data)
	return nil
}
