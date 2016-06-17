// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

var (
	BlacklistTableCols = []string{
		"id",
		"value",
	}

	DeviceTableRows = []string{
		"id",
		"mac",
		"username",
		"registered_from",
		"platform",
		"expires",
		"date_registered",
		"user_agent",
		"blacklisted",
		"description",
		"last_seen",
	}

	LeaseTableCols = []string{
		"id",
		"ip",
		"mac",
		"network",
		"start",
		"end",
		"hostname",
		"abandoned",
		"registered",
	}

	UserTableCols = []string{
		"id",
		"username",
		"password",
		"device_limit",
		"default_expiration",
		"expiration_type",
		"can_manage",
		"can_autoreg",
		"valid_start",
		"valid_end",
		"valid_forever",
	}
)
