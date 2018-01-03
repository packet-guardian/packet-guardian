package tasks

import (
	"os"
	"testing"

	"github.com/packet-guardian/packet-guardian/src/common"
)

func TestFilesystemSessionTaskEmptyDir(t *testing.T) {
	fakeEnv := common.NewEnvironment(common.EnvTesting)
	fakeEnv.Config = common.NewEmptyConfig()
	fakeEnv.Config.Webserver.SessionStore = "filesystem"
	fakeEnv.Config.Webserver.SessionsDir = "sessions"

	os.MkdirAll("sessions", 0755)
	defer os.RemoveAll("sessions")

	_, err := cleanFileSystemSessions(fakeEnv)
	if err != nil {
		t.Fatal(err)
	}
}
