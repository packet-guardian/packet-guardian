// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

// Database table and column names for enumeration and misc use.
var (
	DatabaseTableNames = []string{
		"blacklist",
		"device",
		"lease",
		"lease_history",
		"sessions",
		"settings",
		"user",
	}

	BlacklistTableCols = []string{
		"id",
		"value",
		"comment",
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
		"ui_group",
		"api_group",
		"allow_status_api",
	}

	SessionTableCols = []string{
		"id",
		"session_data",
		"created_on",
		"modified_on",
	}

	SettingTableCols = []string{
		"id",
		"value",
	}
)
