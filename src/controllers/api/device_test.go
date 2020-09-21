package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func editDescriptionTestSetup(duser, user string, perms models.Permission) (*Device, httprouter.Params, *models.Device, *http.Request) {
	e := common.NewTestEnvironment()

	testMac, _ := net.ParseMAC("12:34:56:12:34:56")
	testDeviceStore := &stores.TestDeviceStore{}

	testDevice := models.NewDevice(testDeviceStore, nil, &stores.TestBlacklistItem{Val: false})
	testDevice.MAC = testMac
	testDevice.Description = "Description"
	testDevice.Username = duser

	testDeviceStore.Devices = []*models.Device{testDevice}

	testUserStore := &stores.TestUserStore{}
	testuser := models.NewUser(
		e,
		testUserStore,
		&stores.TestBlacklistItem{Val: false},
		user,
	)
	testuser.Rights = perms
	testUserStore.Users = []*models.User{testuser}

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req = models.SetUserToContext(req, testuser)
	req.PostForm = map[string][]string{
		"description": []string{"edited description"},
	}

	params := []httprouter.Param{
		{
			Key:   "mac",
			Value: "12:34:56:12:34:56",
		},
	}

	return NewDeviceController(e, testUserStore, testDeviceStore, nil), params, testDevice, req
}

func TestDeviceEditDescriptionHandlerSameUser(t *testing.T) {
	testHandler, params, testDevice, req := editDescriptionTestSetup(
		"testuser", "testuser", models.ManageOwnRights)

	w := httptest.NewRecorder()
	testHandler.EditDescriptionHandler(w, req, params)
	if w.Code != 200 {
		t.Errorf("Wrong HTTP code. Expected 200, got %d", w.Code)
	}

	if testDevice.Description != "edited description" {
		t.Errorf("Description wasn't changed. Expected %s, got %s", "edited description", testDevice.Description)
	}
}

func TestDeviceEditDescriptionHandlerDifferentUserNotAdmin(t *testing.T) {
	testHandler, params, testDevice, req := editDescriptionTestSetup(
		"otheruser", "testuser", models.ManageOwnRights)

	w := httptest.NewRecorder()
	testHandler.EditDescriptionHandler(w, req, params)
	if w.Code != 401 {
		t.Errorf("Wrong HTTP code. Expected 401, got %d", w.Code)
	}

	if testDevice.Description != "Description" {
		t.Errorf("Description wasn't changed. Expected %s, got %s", "Description", testDevice.Description)
	}
}

func TestDeviceEditDescriptionHandlerSameUserReadonly(t *testing.T) {
	testHandler, params, testDevice, req := editDescriptionTestSetup(
		"testuser", "testuser", models.ReadOnlyRights.Without(models.ManageOwnRights))

	w := httptest.NewRecorder()
	testHandler.EditDescriptionHandler(w, req, params)
	if w.Code != 401 {
		t.Errorf("Wrong HTTP code. Expected 401, got %d", w.Code)
	}

	if testDevice.Description != "Description" {
		t.Errorf("Description wasn't changed. Expected %s, got %s", "Description", testDevice.Description)
	}
}

func TestDeviceEditDescriptionHandlerDifferentUserReadonly(t *testing.T) {
	testHandler, params, testDevice, req := editDescriptionTestSetup(
		"otheruser", "testuser", models.ReadOnlyRights.Without(models.ManageOwnRights))

	w := httptest.NewRecorder()
	testHandler.EditDescriptionHandler(w, req, params)
	if w.Code != 401 {
		t.Errorf("Wrong HTTP code. Expected 401, got %d", w.Code)
	}

	if testDevice.Description != "Description" {
		t.Errorf("Description wasn't changed. Expected %s, got %s", "Description", testDevice.Description)
	}
}

func TestDeviceEditDescriptionHandlerSameUserAdmin(t *testing.T) {
	testHandler, params, testDevice, req := editDescriptionTestSetup(
		"testuser", "testuser", models.ManageOwnRights.With(models.EditDevice))

	w := httptest.NewRecorder()
	testHandler.EditDescriptionHandler(w, req, params)
	if w.Code != 200 {
		t.Errorf("Wrong HTTP code. Expected 200, got %d", w.Code)
	}

	if testDevice.Description != "edited description" {
		t.Errorf("Description wasn't changed. Expected %s, got %s", "edited description", testDevice.Description)
	}
}

func TestDeviceEditDescriptionHandlerDifferentUserAdmin(t *testing.T) {
	testHandler, params, testDevice, req := editDescriptionTestSetup(
		"otheruser", "testuser", models.ManageOwnRights.With(models.EditDevice))

	w := httptest.NewRecorder()
	testHandler.EditDescriptionHandler(w, req, params)
	if w.Code != 200 {
		t.Errorf("Wrong HTTP code. Expected 200, got %d", w.Code)
	}

	if testDevice.Description != "edited description" {
		t.Errorf("Description wasn't changed. Expected %s, got %s", "edited description", testDevice.Description)
	}
}

type registerTestUser struct {
	username    string
	permissions models.Permission
	blacklisted bool
	delegates   map[string]models.Permission
}

func registrationTestSetup(sessionUser *registerTestUser, otherUsers []*registerTestUser, withLease bool) (*Device, *stores.TestDeviceStore, *http.Request) {
	e := common.NewTestEnvironment()
	e.Config.Registration.AllowManualRegistrations = true

	testDeviceStore := &stores.TestDeviceStore{
		Devices: make([]*models.Device, 0, 1),
	}

	testLeaseStore := &stores.TestLeaseStore{
		Leases: make([]*dhcp.Lease, 0, 1),
	}
	if withLease {
		testLeaseStore.Leases = append(testLeaseStore.Leases, &dhcp.Lease{
			ID:          1,
			IP:          net.ParseIP("10.0.0.1"),
			MAC:         net.HardwareAddr{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef},
			Network:     "test",
			Start:       time.Now().Add(-1 * time.Minute),
			End:         time.Now().Add(time.Minute),
			Hostname:    "The test machine",
			IsAbandoned: false,
			Offered:     false,
			Registered:  false,
		})
	}

	testUserStore := &stores.TestUserStore{}
	storeUsers := make([]*models.User, 0, len(otherUsers)+1)

	sessionuser := models.NewUser(
		e,
		testUserStore,
		&stores.TestBlacklistItem{Val: sessionUser.blacklisted},
		sessionUser.username,
	)
	sessionuser.Rights = sessionUser.permissions
	sessionuser.Delegates = sessionUser.delegates
	storeUsers = append(storeUsers, sessionuser)

	for _, user := range otherUsers {
		testuser := models.NewUser(
			e,
			testUserStore,
			&stores.TestBlacklistItem{Val: user.blacklisted},
			user.username,
		)
		testuser.Rights = user.permissions
		testuser.Delegates = user.delegates
		storeUsers = append(storeUsers, testuser)
	}

	testUserStore.Users = storeUsers

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	req.RemoteAddr = "10.0.0.1:8234"
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req = models.SetUserToContext(req, sessionuser)
	req = common.SetIPToContext(req)

	return NewDeviceController(e, testUserStore, testDeviceStore, testLeaseStore), testDeviceStore, req
}

type registrationTestCase struct {
	testCaseName                                    string
	sessionUser                                     *registerTestUser
	withLease                                       bool
	macAddress                                      string
	formMACAddress, username, platform, description string
	respHTTPCode                                    int
	deviceStoreLen                                  int
	makeDevice                                      bool
	otherUsers                                      []*registerTestUser
}

var registrationTests []registrationTestCase = []registrationTestCase{
	{
		testCaseName: "TestDeviceRegistrationManual",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   200,
		deviceStoreLen: 1,
	},
	{
		testCaseName: "TestDeviceRegistrationManualBlacklisted",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
			blacklisted: true,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   403,
	},
	{
		testCaseName: "TestDeviceRegistrationAutomatic",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
		},
		withLease:      true,
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "",
		username:       "testuser",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   200,
		deviceStoreLen: 1,
	},
	{
		testCaseName: "TestDeviceRegistrationManualOtherUserNonAdmin",
		sessionUser: &registerTestUser{
			username:    "testuser2",
			permissions: models.ManageOwnRights,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   403,
	},
	{
		testCaseName: "TestDeviceRegistrationManualOtherUserGlobalAdmin",
		sessionUser: &registerTestUser{
			username:    "testuser2",
			permissions: models.AdminRights,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   200,
		deviceStoreLen: 1,
	},
	{
		testCaseName: "TestDeviceRegistrationManualAlreadyRegistered",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   409,
		deviceStoreLen: 1,
		makeDevice:     true,
	},
	{
		testCaseName: "TestDeviceRegistrationManualDelegateRW",
		sessionUser: &registerTestUser{
			username:    "delegate",
			permissions: models.ManageOwnRights,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser2",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   200,
		deviceStoreLen: 1,
		otherUsers: []*registerTestUser{
			{
				username:    "testuser2",
				permissions: models.ManageOwnRights,
				delegates: map[string]models.Permission{
					"delegate": models.DelegatePermissions["RW"],
				},
			},
		},
	},
	{
		testCaseName: "TestDeviceRegistrationManualDelegateRO",
		sessionUser: &registerTestUser{
			username:    "delegate",
			permissions: models.ManageOwnRights,
		},
		macAddress:     "12:34:56:ab:cd:ef",
		formMACAddress: "12:34:56:ab:cd:ef",
		username:       "testuser2",
		platform:       "tester",
		description:    "this is a test",
		respHTTPCode:   403,
		deviceStoreLen: 0,
		otherUsers: []*registerTestUser{
			{
				username:    "testuser2",
				permissions: models.ManageOwnRights,
				delegates: map[string]models.Permission{
					"delegate": models.DelegatePermissions["RO"],
				},
			},
		},
	},
}

func TestRegistrationHandler(t *testing.T) {
	for _, testCase := range registrationTests {
		t.Run(testCase.testCaseName, func(t *testing.T) {
			testHandler, devStore, req := registrationTestSetup(testCase.sessionUser, testCase.otherUsers, testCase.withLease)

			if testCase.makeDevice {
				testDevice := models.NewDevice(devStore, nil, &stores.TestBlacklistItem{Val: false})
				devMAC, _ := net.ParseMAC(testCase.macAddress)
				testDevice.ID = 1
				testDevice.MAC = devMAC
				testDevice.Description = testCase.description
				testDevice.Username = testCase.username

				devStore.Devices = append(devStore.Devices, testDevice)
			}

			req.PostForm = map[string][]string{
				"mac-address": {testCase.formMACAddress},
				"username":    {testCase.username},
				"platform":    {testCase.platform},
				"description": {testCase.description},
			}

			w := httptest.NewRecorder()
			testHandler.RegistrationHandler(w, req, nil)
			if w.Code != testCase.respHTTPCode {
				t.Errorf("Wrong HTTP code. Expected %d, got %d", testCase.respHTTPCode, w.Code)
			}

			if len(devStore.Devices) != testCase.deviceStoreLen {
				t.Errorf("Device stores doesn't have the right number of devices: %d, expected %d", len(devStore.Devices), testCase.deviceStoreLen)
			}

			if testCase.deviceStoreLen > 0 && len(devStore.Devices) == testCase.deviceStoreLen {
				device := devStore.Devices[0]
				if device.MAC.String() != testCase.macAddress {
					t.Errorf("Device MAC address incorrect: %s, expected %s", device.MAC.String(), testCase.macAddress)
				}
			}

			if t.Failed() {
				resp, _ := httputil.DumpResponse(w.Result(), true)
				t.Log(string(resp))
			}
		})
	}
}

/// Delete Device Tests
func deleteDeviceTestSetup(sessionUser *registerTestUser, otherUsers []*registerTestUser) (*Device, *stores.TestDeviceStore, *http.Request) {
	e := common.NewTestEnvironment()
	e.Config.Registration.AllowManualRegistrations = true

	testDeviceStore := &stores.TestDeviceStore{
		Devices: make([]*models.Device, 0, 1),
	}

	testLeaseStore := &stores.TestLeaseStore{
		Leases: make([]*dhcp.Lease, 0, 1),
	}

	testUserStore := &stores.TestUserStore{}
	storeUsers := make([]*models.User, 0, len(otherUsers)+1)

	sessionuser := models.NewUser(
		e,
		testUserStore,
		&stores.TestBlacklistItem{Val: sessionUser.blacklisted},
		sessionUser.username,
	)
	sessionuser.Rights = sessionUser.permissions
	sessionuser.Delegates = sessionUser.delegates
	storeUsers = append(storeUsers, sessionuser)

	for _, user := range otherUsers {
		testuser := models.NewUser(
			e,
			testUserStore,
			&stores.TestBlacklistItem{Val: user.blacklisted},
			user.username,
		)
		testuser.Rights = user.permissions
		testuser.Delegates = user.delegates
		storeUsers = append(storeUsers, testuser)
	}

	testUserStore.Users = storeUsers

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	req.RemoteAddr = "10.0.0.1:8234"
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req = models.SetUserToContext(req, sessionuser)
	req = common.SetIPToContext(req)

	return NewDeviceController(e, testUserStore, testDeviceStore, testLeaseStore), testDeviceStore, req
}

type deleteDeviceTestCase struct {
	testCaseName   string
	sessionUser    *registerTestUser
	macs           []string
	username       string
	devicesToMake  []string
	respHTTPCode   int
	deviceStoreLen int
	otherUsers     []*registerTestUser
}

var deleteDeviceTests []deleteDeviceTestCase = []deleteDeviceTestCase{
	{
		testCaseName: "TestDeleteOwnDevice",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
		},
		macs:           []string{"12:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "testuser",
		respHTTPCode:   204,
		deviceStoreLen: 1,
	},
	{
		testCaseName: "TestDeleteOwnDeviceBlacklisted",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
			blacklisted: true,
		},
		macs:           []string{"12:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "testuser",
		respHTTPCode:   403,
		deviceStoreLen: 2,
	},
	{
		testCaseName: "TestDeleteMultipleOwnDevice",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
		},
		macs:           []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "testuser",
		respHTTPCode:   204,
		deviceStoreLen: 0,
	},
	{
		testCaseName: "TestDeleteOtherUserDeviceNonAdmin",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.ManageOwnRights,
		},
		macs:           []string{"12:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "otheruser",
		respHTTPCode:   403,
		deviceStoreLen: 2,
	},
	{
		testCaseName: "TestDeleteOtherUserDeviceGlobalAdmin",
		sessionUser: &registerTestUser{
			username:    "testuser",
			permissions: models.AdminRights,
		},
		macs:           []string{"12:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "otheruser",
		respHTTPCode:   204,
		deviceStoreLen: 1,
	},
	{
		testCaseName: "TestDeleteOtherUserDeviceDelegateRW",
		sessionUser: &registerTestUser{
			username:    "delegate",
			permissions: models.ManageOwnRights,
		},
		macs:           []string{"12:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "otheruser",
		respHTTPCode:   204,
		deviceStoreLen: 1,
		otherUsers: []*registerTestUser{
			{
				username:    "otheruser",
				permissions: models.ManageOwnRights,
				delegates: map[string]models.Permission{
					"delegate": models.DelegatePermissions["RW"],
				},
			},
		},
	},
	{
		testCaseName: "TestDeleteOtherUserDeviceDelegateRO",
		sessionUser: &registerTestUser{
			username:    "delegate",
			permissions: models.ManageOwnRights,
		},
		macs:           []string{"12:34:56:ab:cd:ef"},
		devicesToMake:  []string{"12:34:56:ab:cd:ef", "22:34:56:ab:cd:ef"},
		username:       "otheruser",
		respHTTPCode:   403,
		deviceStoreLen: 2,
		otherUsers: []*registerTestUser{
			{
				username:    "otheruser",
				permissions: models.ManageOwnRights,
				delegates: map[string]models.Permission{
					"delegate": models.DelegatePermissions["RO"],
				},
			},
		},
	},
}

func TestDeleteDeviceHandler(t *testing.T) {
	for _, testCase := range deleteDeviceTests {
		t.Run(testCase.testCaseName, func(t *testing.T) {
			testHandler, devStore, req := deleteDeviceTestSetup(testCase.sessionUser, testCase.otherUsers)

			id := 1
			for _, deviceMAC := range testCase.devicesToMake {
				testDevice := models.NewDevice(devStore, nil, &stores.TestBlacklistItem{Val: false})
				devMAC, _ := net.ParseMAC(deviceMAC)
				testDevice.ID = id
				testDevice.MAC = devMAC
				testDevice.Username = testCase.username
				devStore.Save(testDevice)
				id++
			}

			req.PostForm = map[string][]string{
				"username": {testCase.username},
				"mac":      {strings.Join(testCase.macs, ",")},
			}

			w := httptest.NewRecorder()
			testHandler.DeleteHandler(w, req, nil)
			if w.Code != testCase.respHTTPCode {
				t.Errorf("Wrong HTTP code. Expected %d, got %d", testCase.respHTTPCode, w.Code)
			}

			if len(devStore.Devices) != testCase.deviceStoreLen {
				t.Errorf("Device stores doesn't have the right number of devices: %d, expected %d", len(devStore.Devices), testCase.deviceStoreLen)
			}

			if t.Failed() {
				resp, _ := httputil.DumpResponse(w.Result(), true)
				t.Log(string(resp))
			}
		})
	}
}
