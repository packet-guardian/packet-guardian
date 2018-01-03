// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
)

var sessionExpiration = time.Duration(-24) * time.Hour

func init() {
	RegisterJob("Purge old web sessions", cleanUpExpiredSessions)
}

func cleanUpExpiredSessions(e *common.Environment) (string, error) {
	switch e.Config.Webserver.SessionStore {
	case "filesystem":
		return cleanFileSystemSessions(e)
	case "database":
		return cleanDBSessions(e)
	}
	return "", nil
}

func cleanFileSystemSessions(e *common.Environment) (string, error) {
	w := &sessionWalker{
		n:           time.Now().Add(sessionExpiration),
		sessionsDir: e.Config.Webserver.SessionsDir,
	}
	if err := w.walk(); err != nil {
		return "", err
	}
	return fmt.Sprintf("Deleted %d sessions", w.c), nil
}

type sessionWalker struct {
	n           time.Time
	c           int
	sessionsDir string
}

func (s *sessionWalker) walk() error {
	return filepath.Walk(s.sessionsDir, s.sessionDirWalker)
}

func (s *sessionWalker) sessionDirWalker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() && path != s.sessionsDir {
		return filepath.SkipDir
	}
	if info.ModTime().Before(s.n) {
		s.c++
		return os.Remove(path)
	}
	return nil
}

func cleanDBSessions(e *common.Environment) (string, error) {
	expired := time.Now().Add(sessionExpiration)
	results, err := e.DB.Exec(`DELETE FROM "sessions" WHERE "modified_on" < ?`, expired.Unix())
	if err != nil {
		return "", err
	}
	rowsAffected, _ := results.RowsAffected()
	return fmt.Sprintf("Deleted %d sessions", rowsAffected), nil
}
