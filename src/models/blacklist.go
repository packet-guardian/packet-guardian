// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

var bs *blacklistStore

type blacklistStore struct {
	e *common.Environment
}

func getBlacklistStore(e *common.Environment) *blacklistStore {
	if bs != nil {
		return bs
	}
	bs = &blacklistStore{e: e}
	return bs
}

func (b *blacklistStore) isBlacklisted(s string) bool {
	sql := `SELECT "id" FROM "blacklist" WHERE "value" = ?`
	var id int
	row := b.e.DB.QueryRow(sql, s)
	err := row.Scan(&id)
	return (err == nil)
}

func (b *blacklistStore) addToBlacklist(s string) error {
	sql := `INSERT INTO "blacklist" ("value") VALUES (?)`
	_, err := b.e.DB.Exec(sql, s)
	return err
}

func (b *blacklistStore) removeFromBlacklist(s string) error {
	sql := `DELETE FROM "blacklist" WHERE "value" = ?`
	_, err := b.e.DB.Exec(sql, s)
	return err
}

type blacklistItem struct {
	bs      *blacklistStore
	is      bool
	cached  bool
	changed bool
}

func newBlacklistItem(bs *blacklistStore) *blacklistItem {
	return &blacklistItem{bs: bs}
}

func (b *blacklistItem) blacklist() {
	b.cached = true
	b.is = true
	b.changed = true
}

func (b *blacklistItem) unblacklist() {
	b.cached = true
	b.is = false
	b.changed = true
}

func (b *blacklistItem) isBlacklisted(key string) bool {
	if b.cached {
		return b.is
	}

	b.is = b.bs.isBlacklisted(key)
	b.cached = true
	b.changed = false
	return b.is
}

func (b *blacklistItem) save(key string) error {
	// We only need to do something if the blacklist setting was changed
	if !b.changed {
		return nil
	}

	// If blacklisted, insert into database
	if b.is {
		return b.bs.addToBlacklist(key)
	}

	// Otherwise remove them from the blacklist
	return b.bs.removeFromBlacklist(key)
}
