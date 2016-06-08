// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

const testSomethingHash = "$2a$04$zxGo0fl3SeyWAix1MrxqI.qEgO42Jqx94eAaXtUfqr.SK/pSZBEq2"

func TestLocalAuth(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id",
		"username",
		"password",
		"device_limit",
		"default_expiration",
		"expiration_type",
		"can_manage",
		"valid_forever",
		"valid_start",
		"valid_end",
	}).AddRow(
		1,
		"tester1",
		testSomethingHash,
		0, 0, 0, 1, 1, 0, 0,
	)

	passRows := sqlmock.NewRows([]string{"password"}).AddRow(testSomethingHash)

	mock.ExpectQuery("SELECT .*? FROM \"user\"").
		WithArgs("tester1").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT \"password\"").
		WithArgs("tester1").
		WillReturnRows(passRows)

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	common.SetEnvironmentToContext(req, e)
	common.SetSessionToContext(req, session)

	req.Form = make(url.Values)
	req.Form.Add("username", "tester1")
	req.Form.Add("password", "testSomething")

	local := &localAuthenticator{}

	if !local.loginUser(req, httptest.NewRecorder()) {
		t.Error("Failed to login user. Expected true, got false")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestFailedLocalAuth(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id",
		"username",
		"password",
		"device_limit",
		"default_expiration",
		"expiration_type",
		"can_manage",
		"valid_forever",
		"valid_start",
		"valid_end",
	}).AddRow(
		1,
		"tester1",
		testSomethingHash,
		0, 0, 0, 1, 1, 0, 0,
	)

	passRows := sqlmock.NewRows([]string{"password"}).AddRow(testSomethingHash)

	mock.ExpectQuery("SELECT .*? FROM \"user\"").
		WithArgs("tester1").
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT \"password\"").
		WithArgs("tester1").
		WillReturnRows(passRows)

	e := common.NewTestEnvironment()
	e.DB = &common.DatabaseAccessor{DB: db}

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	common.SetEnvironmentToContext(req, e)
	common.SetSessionToContext(req, session)

	req.Form = make(url.Values)
	req.Form.Add("username", "tester1")
	req.Form.Add("password", "testSomething1") // Incorrect password

	local := &localAuthenticator{}

	if local.loginUser(req, httptest.NewRecorder()) {
		t.Error("Failed to login user. Expected false, got true")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
