package api

import (
	"testing"

	"net"
	"net/http"
	"net/http/httptest"

	"github.com/julienschmidt/httprouter"
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
