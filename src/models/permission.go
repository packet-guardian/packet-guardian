// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// Permission is an unsigned int where each bit represents an individual permission.
type Permission uint64

const (
	// These four are administrative permissions for user management
	ViewUsers Permission = 1 << iota
	CreateUser
	EditUser
	DeleteUser

	// These are administrative permissions for managing devices other than
	// the current user.
	ViewDevices
	CreateDevice
	EditDevice
	DeleteDevice
	ReassignDevice

	// These are normal user permissions to manage their own devices.
	ViewOwn
	CreateOwn  // Manually register device
	AutoRegOwn // Automatically register device
	EditOwn
	DeleteOwn

	// Administrative permissions for blacklist
	ManageBlacklist
	BypassBlacklist

	// Flag to view the admin dashboard. This should be given to any group
	// that has at lease ViewUsers or ViewDevices
	ViewAdminPage
)

const (
	// AdminRights has all bits set to one meaning all permissions are given
	AdminRights = 1<<64 - 1
	// HelpDeskRights represents a restricted admin user
	HelpDeskRights = ReadOnlyRights |
		ViewUsers |
		CreateDevice |
		EditDevice |
		DeleteDevice
	// ReadOnlyRights represents a read-only admin user
	ReadOnlyRights = ViewOwn |
		ManageOwnRights |
		ViewDevices |
		ViewAdminPage
	// ManageOwnRights is a convenience Permission combining CreateOwn, EditOwn and DeleteOwn.
	ManageOwnRights = CreateOwn |
		AutoRegOwn |
		EditOwn |
		DeleteOwn
)

// With returns a new Permission where p now has permission(s) new.
func (p Permission) With(new Permission) Permission {
	return p | new
}

// Can checks if p is 1 for ALL permission(s) represented by check.
func (p Permission) Can(check Permission) bool {
	return (p&check == check)
}

// CanEither checks if p is 1 for ANY permission(s) represented by check.
// CanEither will return true if any permission in check is present.
func (p Permission) CanEither(check Permission) bool {
	return (p&check != 0)
}

// Without will return a new Permission where rm is removed from p.
func (p Permission) Without(rm Permission) Permission {
	return p &^ rm
}
