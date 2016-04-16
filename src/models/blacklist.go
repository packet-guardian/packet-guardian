package models

import "github.com/onesimus-systems/packet-guardian/src/common"

type BlacklistItem struct {
	ID    int
	Value string
}

func GetAllBlacklistedItems(e *common.Environment) ([]*BlacklistItem, error) {
	return nil, nil
}
