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

func init() {
	RegisterJob("Purge old web sessions", cleanUpExpiredSessions)
}

func cleanUpExpiredSessions(e *common.Environment) (string, error) {
	w := sessionWalker{n: time.Now().Add(time.Duration(-24) * time.Hour)}
	err := filepath.Walk(sessionsDir, w.sessionDirWalker)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Purged %d sessions", w.c), nil
}

type sessionWalker struct {
	n time.Time
	c int
}

func (s sessionWalker) sessionDirWalker(path string, info os.FileInfo, err error) error {
	if info.IsDir() && path != sessionsDir {
		return filepath.SkipDir
	}
	if info.ModTime().Before(s.n) {
		s.c++
		return os.Remove(path)
	}
	return nil
}
