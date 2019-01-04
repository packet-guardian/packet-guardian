package stats

import (
	"github.com/lfkeitel/verbose/v4"
	dhcp "github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
)

func DHCPNetworkList(e *common.Environment) []string {
	dhcpConfig, err := dhcp.ParseFile(e.Config.DHCP.ConfigFile)
	if err != nil {
		e.Log.WithFields(verbose.Fields{
			"error":   err,
			"package": "reports:dhcp-pools",
		}).Error("Error loading DHCP configuration")
		return nil
	}

	return dhcpConfig.Networks()
}
