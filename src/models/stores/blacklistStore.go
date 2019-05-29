// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stores

import (
	"github.com/packet-guardian/packet-guardian/src/common"
)

var appBlacklistStore BlacklistStore

type BlacklistStore interface {
	IsBlacklisted(s string) bool
	AddToBlacklist(s string) error
	RemoveFromBlacklist(s string) error
}

type blacklistStore struct {
	e *common.Environment
}

func newBlacklistStore(e *common.Environment) *blacklistStore {
	return &blacklistStore{
		e: e,
	}
}

func GetBlacklistStore(e *common.Environment) BlacklistStore {
	if appBlacklistStore == nil {
		appBlacklistStore = newBlacklistStore(e)
	}
	return appBlacklistStore
}

func (b *blacklistStore) IsBlacklisted(s string) bool {
	if s == "" {
		return false
	}

	sql := `SELECT "id" FROM "blacklist" WHERE "value" = ?`
	if b.e.DB == nil {
		b.e.Log.Alert("Database is nil in blacklist store")
		return false
	}
	var id int
	row := b.e.DB.QueryRow(sql, s)
	err := row.Scan(&id)
	return (err == nil)
}

func (b *blacklistStore) AddToBlacklist(s string) error {
	if s == "" {
		return nil
	}

	sql := `INSERT INTO "blacklist" ("value") VALUES (?)`
	_, err := b.e.DB.Exec(sql, s)
	return err
}

func (b *blacklistStore) RemoveFromBlacklist(s string) error {
	if s == "" {
		return nil
	}

	sql := `DELETE FROM "blacklist" WHERE "value" = ?`
	_, err := b.e.DB.Exec(sql, s)
	return err
}

type BlacklistItem interface {
	Blacklist()
	Unblacklist()
	IsBlacklisted(string) bool
	Save(string) error
}

type blacklistItem struct {
	bs      BlacklistStore
	is      bool
	cached  bool
	changed bool
}

func NewBlacklistItem(bs BlacklistStore) BlacklistItem {
	return &blacklistItem{bs: bs}
}

func (b *blacklistItem) Blacklist() {
	b.cached = true
	b.is = true
	b.changed = true
}

func (b *blacklistItem) Unblacklist() {
	b.cached = true
	b.is = false
	b.changed = true
}

func (b *blacklistItem) IsBlacklisted(key string) bool {
	if b.cached {
		return b.is
	}

	b.is = b.bs.IsBlacklisted(key)
	b.cached = true
	b.changed = false
	return b.is
}

func (b *blacklistItem) Save(key string) error {
	// We only need to do something if the blacklist setting was changed
	if !b.changed {
		return nil
	}

	// If blacklisted, insert into database
	if b.is {
		return b.bs.AddToBlacklist(key)
	}

	// Otherwise remove them from the blacklist
	return b.bs.RemoveFromBlacklist(key)
}
