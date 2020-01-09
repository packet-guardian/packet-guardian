// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

func TestRootHandlerLoggedInNormal(t *testing.T) {
	e := common.NewTestEnvironment()

	session := common.NewTestSession()
	session.Set("loggedin", true)
	session.Set("username", "testUser")

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)

	sessionUser := models.NewUser(e, &stores.TestUserStore{}, &stores.TestBlacklistItem{}, "testUser")
	req = models.SetUserToContext(req, sessionUser)

	w := httptest.NewRecorder()
	(&rootHandler{}).ServeHTTP(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/manage" {
		t.Errorf("Wrong location. Expected /manage, got %s", w.HeaderMap.Get("Location"))
	}
}

func TestRootHandlerLoggedInAdmin(t *testing.T) {
	e := common.NewTestEnvironment()

	session := common.NewTestSession()
	session.Set("loggedin", true)
	session.Set("username", "testUser")

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)

	sessionUser := models.NewUser(e, &stores.TestUserStore{}, &stores.TestBlacklistItem{}, "testUser")
	sessionUser.UIGroup = "admin"
	sessionUser.LoadRights()

	req = models.SetUserToContext(req, sessionUser)

	w := httptest.NewRecorder()
	(&rootHandler{}).ServeHTTP(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/admin" {
		t.Errorf("Wrong location. Expected /admin, got %s", w.HeaderMap.Get("Location"))
	}
}

func TestRootHandlerNotLoggedInNotRegistered(t *testing.T) {
	testMac, _ := net.ParseMAC("12:34:56:12:34:56")
	testLease := &dhcp.Lease{
		ID:  1,
		IP:  net.ParseIP("192.168.1.10"),
		MAC: testMac,
	}

	testLeaseStore := &stores.TestLeaseStore{
		Leases: []*dhcp.Lease{testLease},
	}

	e := common.NewTestEnvironment()

	session := common.NewTestSession()
	session.Set("loggedin", false)

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req.RemoteAddr = "192.168.1.10"
	req = common.SetIPToContext(req)

	w := httptest.NewRecorder()
	(&rootHandler{leases: testLeaseStore}).ServeHTTP(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/register" {
		t.Errorf("Wrong location. Expected /register, got %s", w.HeaderMap.Get("Location"))
	}
}

func TestRootHandlerNotLoggedInRegistered(t *testing.T) {
	testMac, _ := net.ParseMAC("12:34:56:12:34:56")
	testLease := &dhcp.Lease{
		ID:         1,
		IP:         net.ParseIP("192.168.1.10"),
		MAC:        testMac,
		Registered: true,
	}

	testLeaseStore := &stores.TestLeaseStore{
		Leases: []*dhcp.Lease{testLease},
	}

	e := common.NewTestEnvironment()

	session := common.NewTestSession()
	session.Set("loggedin", false)

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req.RemoteAddr = "192.168.1.10"
	req = common.SetIPToContext(req)

	w := httptest.NewRecorder()
	(&rootHandler{leases: testLeaseStore}).ServeHTTP(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/login" {
		t.Errorf("Wrong location. Expected /login, got %s", w.HeaderMap.Get("Location"))
	}
}
