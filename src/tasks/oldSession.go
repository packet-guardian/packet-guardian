// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

const sessionsDir string = "sessions"

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
	w := &sessionWalker{n: time.Now().Add(sessionExpiration)}
	err := filepath.Walk(sessionsDir, w.sessionDirWalker)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Deleted %d sessions", w.c), nil
}

type sessionWalker struct {
	n time.Time
	c int
}

func (s *sessionWalker) sessionDirWalker(path string, info os.FileInfo, err error) error {
	if info.IsDir() && path != sessionsDir {
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
