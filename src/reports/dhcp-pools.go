package reports

import (
	"bytes"
	"net"
	"net/http"
	"sort"

	"github.com/lfkeitel/verbose/v4"
	dhcp "github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func init() {
	RegisterReport("dhcp-pools", "DHCP Pool Statistics", poolReport)
}

func poolReport(e *common.Environment, w http.ResponseWriter, r *http.Request, stores stores.StoreCollection) error {
	dhcpConfig, err := dhcp.ParseFile(e.Config.DHCP.ConfigFile)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:dhcp-pools",
		}).Error("Error loading DHCP configuration")
		return err
	}

	dhcpPkgConfig := &dhcp.ServerConfig{
		LeaseStore: stores.Leases,
	}

	handler := dhcp.NewDHCPServer(dhcpConfig, dhcpPkgConfig)
	handler.LoadLeases()
	stats := handler.GetPoolStats()

	sort.Stable(poolSubnetSorter(stats))
	sort.Stable(poolNameSorter(stats))

	data := map[string]interface{}{
		"pools": stats,
	}

	e.Views.NewView("reports-network-pools", r).Render(w, data)
	return nil
}

type poolNameSorter []*dhcp.PoolStat

func (l poolNameSorter) Len() int           { return len(l) }
func (l poolNameSorter) Less(i, j int) bool { return l[i].NetworkName < l[j].NetworkName }
func (l poolNameSorter) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

type poolSubnetSorter []*dhcp.PoolStat

func (l poolSubnetSorter) Len() int { return len(l) }
func (l poolSubnetSorter) Less(i, j int) bool {
	ipI, _, _ := net.ParseCIDR(l[i].Subnet)
	ipJ, _, _ := net.ParseCIDR(l[j].Subnet)
	return bytes.Compare([]byte(ipI), []byte(ipJ)) < 0
}
func (l poolSubnetSorter) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
