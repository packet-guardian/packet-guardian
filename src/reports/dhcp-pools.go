package reports

import (
	"encoding/json"
	"net/http"

	"github.com/lfkeitel/verbose"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models/stores"
	"github.com/usi-lfkeitel/pg-dhcp"
)

func init() {
	RegisterReport("dhcp-pools", "DHCP Pool Statistics", poolReport)
}

func poolReport(e *common.Environment, w http.ResponseWriter, r *http.Request) error {
	dhcpConfig, err := dhcp.ParseFile(e.Config.DHCP.ConfigFile)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:dhcp-pools",
		}).Error("Error loading DHCP configuration")
		return err
	}

	dhcpPkgConfig := &dhcp.ServerConfig{
		LeaseStore: stores.GetLeaseStore(e),
	}

	handler := dhcp.NewDHCPServer(dhcpConfig, dhcpPkgConfig)
	handler.LoadLeases()
	stats := handler.GetPoolStats()
	b, _ := json.MarshalIndent(stats, "", "  ")
	w.Write(b)
	return nil
}
