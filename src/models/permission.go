// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "bytes"

// Permission is an unsigned int where each bit represents an individual permission.
type Permission uint64

const (
	// These four are administrative permissions for user management
	ViewUsers Permission = 1 << iota
	CreateUser
	EditUser
	DeleteUser
	EditUserPermissions

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

	// Allow an account to login even in guest mode
	BypassGuestLogin

	// Flag to view the admin dashboard. This should be given to any group
	// that has at lease ViewUsers or ViewDevices
	ViewAdminPage
	ViewReports

	ViewDebugInfo

	APIRead
	APIWrite
)

const (
	// AdminRights has all bits set to one meaning all permissions are given
	AdminRights Permission = 1<<64 - 1
	// HelpDeskRights represents a restricted admin user
	HelpDeskRights Permission = ReadOnlyRights |
		ViewUsers |
		CreateDevice |
		EditDevice |
		DeleteDevice
	// ReadOnlyRights represents a read-only admin user
	ReadOnlyRights Permission = ViewOwn |
		ManageOwnRights |
		ViewDevices |
		ViewAdminPage |
		ViewReports |
		BypassGuestLogin
	// ManageOwnRights is a convenience Permission combining CreateOwn, EditOwn and DeleteOwn.
	ManageOwnRights Permission = CreateOwn |
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

var uiPermissions = map[string]Permission{
	"admin":    AdminRights.Without(APIRead).Without(APIWrite),
	"helpdesk": HelpDeskRights,
	"readonly": ReadOnlyRights,
}

var apiPermissions = map[string]Permission{
	"readonly-api":  APIRead,
	"readwrite-api": APIRead.With(APIWrite),
	"status-api":    ViewDebugInfo,
}

// DelegatePermissions maps a permissions string to the Permission model
var DelegatePermissions = map[string]Permission{
	"RW": ViewDevices | CreateDevice | EditDevice | DeleteDevice,
	"RO": ViewDevices,
}

// DelegateName returns the name of a delegate permission
func (p Permission) DelegateName() string {
	if p == ViewDevices {
		return "RO"
	}
	return "RW"
}

func (p Permission) String() string {
	buf := bytes.Buffer{}

	if p.Can(ViewUsers) {
		buf.WriteString("models.ViewUsers\n")
	}
	if p.Can(CreateUser) {
		buf.WriteString("models.CreateUser\n")
	}
	if p.Can(EditUser) {
		buf.WriteString("models.EditUser\n")
	}
	if p.Can(DeleteUser) {
		buf.WriteString("models.DeleteUser\n")
	}
	if p.Can(EditUserPermissions) {
		buf.WriteString("models.EditUserPermissions\n")
	}
	if p.Can(ViewDevices) {
		buf.WriteString("models.ViewDevices\n")
	}
	if p.Can(CreateDevice) {
		buf.WriteString("models.CreateDevice\n")
	}
	if p.Can(EditDevice) {
		buf.WriteString("models.EditDevice\n")
	}
	if p.Can(DeleteDevice) {
		buf.WriteString("models.DeleteDevice\n")
	}
	if p.Can(ReassignDevice) {
		buf.WriteString("models.ReassignDevice\n")
	}
	if p.Can(ViewOwn) {
		buf.WriteString("models.ViewOwn\n")
	}
	if p.Can(CreateOwn) {
		buf.WriteString("models.CreateOwn\n")
	}
	if p.Can(AutoRegOwn) {
		buf.WriteString("models.AutoRegOwn\n")
	}
	if p.Can(EditOwn) {
		buf.WriteString("models.EditOwn\n")
	}
	if p.Can(DeleteOwn) {
		buf.WriteString("models.DeleteOwn\n")
	}
	if p.Can(ManageBlacklist) {
		buf.WriteString("models.ManageBlacklist\n")
	}
	if p.Can(BypassBlacklist) {
		buf.WriteString("models.BypassBlacklist\n")
	}
	if p.Can(BypassGuestLogin) {
		buf.WriteString("models.BypassGuestLogin\n")
	}
	if p.Can(ViewAdminPage) {
		buf.WriteString("models.ViewAdminPage\n")
	}
	if p.Can(ViewReports) {
		buf.WriteString("models.ViewReports\n")
	}
	if p.Can(ViewDebugInfo) {
		buf.WriteString("models.ViewDebugInfo\n")
	}
	if p.Can(APIRead) {
		buf.WriteString("models.APIRead\n")
	}
	if p.Can(APIWrite) {
		buf.WriteString("models.APIWrite\n")
	}

	return buf.String()
}
