// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

func TestRootHandlerLoggedInNormal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRow := sqlmock.NewRows(common.UserTableCols).AddRow(
		1, "testUser", "", 0, 0, 0, 1, 1, 1, 0, 0,
	)

	mock.ExpectQuery("SELECT .*? FROM \"user\"").WithArgs("testuser").WillReturnRows(userRow)

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}

	session := common.NewTestSession()
	session.Set("loggedin", true)
	session.Set("username", "testUser")

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)

	sessionUser, err := models.GetUserByUsername(e, "testUser")
	if err != nil {
		t.Logf("Failed to get session user: %v", err.Error())
	}
	req = models.SetUserToContext(req, sessionUser)

	w := httptest.NewRecorder()
	rootHandler(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/manage" {
		t.Errorf("Wrong location. Expected /manage, got %s", w.HeaderMap.Get("Location"))
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestRootHandlerLoggedInAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRow := sqlmock.NewRows(common.UserTableCols).AddRow(
		1, "testUser", "", 0, 0, 0, 1, 1, 1, 0, 0,
	)

	mock.ExpectQuery("SELECT .*? FROM \"user\"").WithArgs("testuser").WillReturnRows(userRow)

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}
	e.Config.Auth.AdminUsers = []string{"testUser"}

	session := common.NewTestSession()
	session.Set("loggedin", true)
	session.Set("username", "testUser")

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)

	sessionUser, err := models.GetUserByUsername(e, "testUser")
	if err != nil {
		t.Logf("Failed to get session user: %v", err.Error())
	}
	req = models.SetUserToContext(req, sessionUser)

	w := httptest.NewRecorder()
	rootHandler(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/admin" {
		t.Errorf("Wrong location. Expected /admin, got %s", w.HeaderMap.Get("Location"))
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestRootHandlerNotLoggedInNotRegistered(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	leaseRow := sqlmock.NewRows(common.LeaseTableCols).AddRow(
		1, "192.168.1.10", "12:34:56:12:34:56", "", 0, 0, "", 0, 0,
	)

	mock.ExpectQuery("SELECT .*? FROM \"lease\"").WithArgs("192.168.1.10").WillReturnRows(leaseRow)

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}

	session := common.NewTestSession()
	session.Set("loggedin", false)

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req.RemoteAddr = "192.168.1.10"
	req = common.SetIPToContext(req)

	w := httptest.NewRecorder()
	rootHandler(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/register" {
		t.Errorf("Wrong location. Expected /register, got %s", w.HeaderMap.Get("Location"))
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestRootHandlerNotLoggedInRegistered(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	leaseRow := sqlmock.NewRows(common.LeaseTableCols).AddRow(
		1, "192.168.1.10", "12:34:56:12:34:56", "", 0, 0, "", 0, 1,
	)

	mock.ExpectQuery("SELECT .*? FROM \"lease\"").WithArgs("192.168.1.10").WillReturnRows(leaseRow)

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}

	session := common.NewTestSession()
	session.Set("loggedin", false)

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)
	req.RemoteAddr = "192.168.1.10"
	req = common.SetIPToContext(req)

	w := httptest.NewRecorder()
	rootHandler(w, req)
	if w.Code != 307 {
		t.Errorf("Wrong redirect code. Expected 307, got %d", w.Code)
	}
	if w.HeaderMap.Get("Location") != "/login" {
		t.Errorf("Wrong location. Expected /login, got %s", w.HeaderMap.Get("Location"))
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
