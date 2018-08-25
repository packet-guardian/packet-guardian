// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import "time"

type PoolStat struct {
	NetworkName, Subnet                     string
	Start                                   string
	End                                     string
	Registered                              bool
	Total, Active, Claimed, Abandoned, Free int
}

func (h *Handler) GetPoolStats() []*PoolStat {
	stats := make([]*PoolStat, 0)
	now := time.Now()
	regFreeTime := time.Duration(c.global.registeredSettings.freeLeaseAfter) * time.Second
	unRegFreeTime := time.Duration(c.global.unregisteredSettings.freeLeaseAfter) * time.Second

	for _, n := range c.networks {
		for _, s := range n.subnets {
			for _, p := range s.pools {
				ps := &PoolStat{
					NetworkName: n.name,
					Subnet:      s.net.String(),
					Registered:  !s.allowUnknown,
					Total:       p.getCountOfIPs(),
					Start:       p.rangeStart.String(),
					End:         p.rangeEnd.String(),
				}

				for _, l := range p.leases {
					if l.IsAbandoned {
						ps.Abandoned++
						continue
					}
					if !l.IsExpired() {
						ps.Active++
						continue
					}
					if !l.Registered && l.End.Add(unRegFreeTime).After(now) { // Unregisted lease expired
						ps.Claimed++
						continue
					}
					if l.Registered && l.End.Add(regFreeTime).After(now) { // Registered lease expired
						ps.Claimed++
						continue
					}
					if l.IsFree() {
						ps.Free++
						continue
					}
				}

				stats = append(stats, ps)
			}
		}
	}
	return stats
}
