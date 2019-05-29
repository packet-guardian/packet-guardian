// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"testing"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

const testSomethingHash = "$2a$04$zxGo0fl3SeyWAix1MrxqI.qEgO42Jqx94eAaXtUfqr.SK/pSZBEq2"

func TestLocalAuth(t *testing.T) {
	testUser := &models.User{
		ID:           1,
		Username:     "tester1",
		Password:     testSomethingHash,
		ValidForever: true,
	}

	testUserStore := &stores.TestUserStore{
		Users: []*models.User{testUser},
	}

	e := common.NewTestEnvironment()

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)

	local := &localAuthenticator{}

	if !local.checkLogin("tester1", "testSomething", req, testUserStore) {
		t.Error("Failed to login user. Expected true, got false")
	}
}

func TestFailedLocalAuth(t *testing.T) {
	testUser := &models.User{
		ID:           1,
		Username:     "tester1",
		Password:     testSomethingHash,
		ValidForever: true,
	}

	testUserStore := &stores.TestUserStore{
		Users: []*models.User{testUser},
	}

	e := common.NewTestEnvironment()

	session := common.NewTestSession()

	req, _ := http.NewRequest("", "", nil)
	req = common.SetEnvironmentToContext(req, e)
	req = common.SetSessionToContext(req, session)

	local := &localAuthenticator{}

	if local.checkLogin("tester1", "testSomething1", req, testUserStore) {
		t.Error("Failed to login user. Expected false, got true")
	}
}
