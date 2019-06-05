package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
)

var nullHTTPHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

type TestBlacklistItem struct{ val bool }

func newTestBlacklistItem(v bool) *TestBlacklistItem   { return &TestBlacklistItem{v} }
func (b *TestBlacklistItem) Blacklist()                { b.val = true }
func (b *TestBlacklistItem) Unblacklist()              { b.val = false }
func (b *TestBlacklistItem) IsBlacklisted(string) bool { return b.val }
func (b *TestBlacklistItem) Save(string) error         { return nil }

func TestCheckAdminMiddleware(t *testing.T) {
	testuser := models.NewUser(common.NewTestEnvironment(), nil, newTestBlacklistItem(false), "testuser")

	req, _ := http.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	adminHandler := CheckAdmin(nullHTTPHandler)
	req = models.SetUserToContext(req, testuser)

	// No admin rights
	testWriter := httptest.NewRecorder()
	adminHandler.ServeHTTP(testWriter, req)
	if testWriter.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected %d, got %d", http.StatusTemporaryRedirect, testWriter.Code)
	}

	// Can view non-special admin pages
	testuser.Rights = testuser.Rights.With(models.ViewAdminPage)
	testWriter = httptest.NewRecorder()
	adminHandler.ServeHTTP(testWriter, req)
	if testWriter.Code != http.StatusOK {
		t.Fatalf("Expected %d, got %d", http.StatusOK, testWriter.Code)
	}

	// Can't view users page
	req.URL.Path = "/admin/users"
	testWriter = httptest.NewRecorder()
	adminHandler.ServeHTTP(testWriter, req)
	if testWriter.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected %d, got %d", http.StatusTemporaryRedirect, testWriter.Code)
	}

	// Can view users page
	testuser.Rights = testuser.Rights.With(models.ViewUsers)
	testWriter = httptest.NewRecorder()
	adminHandler.ServeHTTP(testWriter, req)
	if testWriter.Code != http.StatusOK {
		t.Fatalf("Expected %d, got %d", http.StatusOK, testWriter.Code)
	}
}
