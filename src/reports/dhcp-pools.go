package reports

import (
	"encoding/json"
	"net/http"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
	"github.com/usi-lfkeitel/pg-dhcp"
)

func init() {
	RegisterReport("dhcp-pools", "DHCP Pool Statistics", poolReport)
}

func poolReport(e *common.Environment, w http.ResponseWriter, r *http.Request) error {
	dhcpConfig, err := dhcp.ParseFile(e.Config.DHCP.ConfigFile)
	if err != nil {
		e.Log.WithField("ErrMsg", err).Fatal("Error loading DHCP configuration")
		return err
	}

	dhcpPkgConfig := &dhcp.ServerConfig{
		LeaseStore: models.NewLeaseStore(e),
	}

	handler := dhcp.NewDHCPServer(dhcpConfig, dhcpPkgConfig)
	handler.LoadLeases()
	stats := handler.GetPoolStats()
	b, _ := json.MarshalIndent(stats, "", "  ")
	w.Write(b)
	return nil
}
