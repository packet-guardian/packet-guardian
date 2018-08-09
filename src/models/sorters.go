package models

import (
	"bytes"

	"github.com/packet-guardian/dhcp-lib"
)

// LeaseSorter sorts a slice of leases by their IP addresses
type LeaseSorter []*dhcp.Lease

func (l LeaseSorter) Len() int           { return len(l) }
func (l LeaseSorter) Less(i, j int) bool { return bytes.Compare([]byte(l[i].IP), []byte(l[j].IP)) < 0 }
func (l LeaseSorter) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// UsernameSorter sorts a slice of users by their usernames
type UsernameSorter []*User

func (u UsernameSorter) Len() int           { return len(u) }
func (u UsernameSorter) Less(i, j int) bool { return u[i].Username < u[j].Username }
func (u UsernameSorter) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

// MACSorter sorts a slice of devices by their MAC addresses
type MACSorter []*Device

func (m MACSorter) Len() int           { return len(m) }
func (m MACSorter) Less(i, j int) bool { return bytes.Compare([]byte(m[i].MAC), []byte(m[j].MAC)) < 0 }
func (m MACSorter) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
