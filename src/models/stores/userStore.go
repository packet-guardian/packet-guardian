// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stores

import (
	"errors"
	"strings"
	"time"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
	"github.com/usi-lfkeitel/packet-guardian/src/models"
)

var userStore *UserStore

type UserStore struct {
	e *common.Environment
}

func NewUserStore(e *common.Environment) *UserStore {
	return &UserStore{
		e: e,
	}
}

func GetUserStore(e *common.Environment) *UserStore {
	if userStore == nil || e.IsTesting() {
		userStore = NewUserStore(e)
	}
	return userStore
}

func (s *UserStore) GetUserByUsername(username string) (*models.User, error) {
	if username == "" {
		return models.NewUser(s.e, s, NewBlacklistItem(GetBlacklistStore(s.e))), nil
	}

	username = strings.ToLower(username)

	sql := `WHERE "username" = ?`
	users, err := s.getUsersFromDatabase(sql, username)
	if len(users) == 0 {
		u := models.NewUser(s.e, s, NewBlacklistItem(GetBlacklistStore(s.e)))
		u.Username = username
		u.LoadRights()
		return u, err
	}
	users[0].LoadRights()
	return users[0], nil
}

func (s *UserStore) GetAllUsers() ([]*models.User, error) {
	sql := `ORDER BY "username"`
	if s.e.DB.Driver == "sqlite" {
		sql += " COLLATE NOCASE"
	}
	sql += " ASC"
	return s.getUsersFromDatabase(sql)
}

func (s *UserStore) SearchUsersByField(field, pattern string) ([]*models.User, error) {
	sql := `WHERE "` + field + `" LIKE ?`
	return s.getUsersFromDatabase(sql, pattern)
}

func (s *UserStore) getUsersFromDatabase(where string, values ...interface{}) ([]*models.User, error) {
	sql := `SELECT "id", "username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "can_autoreg", "valid_forever", "valid_start", "valid_end" FROM "user" ` + where
	rows, err := s.e.DB.Query(sql, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.User
	for rows.Next() {
		var id int
		var username string
		var password string
		var deviceLimit int
		var defaultExpiration int64
		var expirationType int
		var canManage bool
		var canAutoreg bool
		var validForever bool
		var validStart int64
		var validEnd int64

		err := rows.Scan(
			&id,
			&username,
			&password,
			&deviceLimit,
			&defaultExpiration,
			&expirationType,
			&canManage,
			&canAutoreg,
			&validForever,
			&validStart,
			&validEnd,
		)
		if err != nil {
			continue
		}

		user := models.NewUser(s.e, s, NewBlacklistItem(GetBlacklistStore(s.e)))
		user.ID = id
		user.Username = username
		user.HasPassword = (password != "")
		user.DeviceLimit = models.UserDeviceLimit(deviceLimit)
		user.ValidStart = time.Unix(validStart, 0)
		user.ValidEnd = time.Unix(validEnd, 0)
		user.ValidForever = validForever
		user.CanManage = canManage
		user.CanAutoreg = canAutoreg
		user.Rights = models.ViewOwn

		if canManage {
			user.Rights = user.Rights.With(models.ManageOwnRights)
		}
		if canAutoreg {
			user.Rights = user.Rights.With(models.AutoRegOwn)
		}
		user.DeviceExpiration = &models.UserDeviceExpiration{
			Mode:  models.UserExpiration(expirationType),
			Value: defaultExpiration,
		}
		results = append(results, user)
	}
	return results, nil
}

func (s *UserStore) GetPassword(username string) (string, error) {
	result := s.e.DB.QueryRow(`SELECT "password" FROM "user" WHERE "username" = ?`, username)
	var p string
	err := result.Scan(&p)
	return p, err
}

func (s *UserStore) Save(u *models.User) error {
	if u.ID == 0 {
		return s.saveNew(u)
	}
	return s.updateExisting(u)
}

func (s *UserStore) updateExisting(u *models.User) error {
	sql := `UPDATE "user" SET "device_limit" = ?, "default_expiration" = ?, "expiration_type" = ?, "can_manage" = ?, "can_autoreg" = ?, "valid_forever" = ?, "valid_start" = ?, "valid_end" = ?`

	if u.NeedToSavePassword() {
		sql += ", \"password\" = ?"
	}

	sql += " WHERE \"id\" = ?"

	var err error
	if u.NeedToSavePassword() {
		_, err = s.e.DB.Exec(
			sql,
			u.DeviceLimit,
			u.DeviceExpiration.Value,
			u.DeviceExpiration.Mode,
			u.CanManage,
			u.CanAutoreg,
			u.ValidForever,
			u.ValidStart.Unix(),
			u.ValidEnd.Unix(),
			u.Password,
			u.ID,
		)
	} else {
		_, err = s.e.DB.Exec(
			sql,
			u.DeviceLimit,
			u.DeviceExpiration.Value,
			u.DeviceExpiration.Mode,
			u.CanManage,
			u.CanAutoreg,
			u.ValidForever,
			u.ValidStart.Unix(),
			u.ValidEnd.Unix(),
			u.ID,
		)
	}
	return err
}

func (s *UserStore) saveNew(u *models.User) error {
	if u.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := `INSERT INTO "user" ("username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "can_autoreg", "valid_forever", "valid_start", "valid_end") VALUES (?,?,?,?,?,?,?,?,?,?)`

	result, err := s.e.DB.Exec(
		sql,
		u.Username,
		u.Password,
		u.DeviceLimit,
		u.DeviceExpiration.Value,
		u.DeviceExpiration.Mode,
		u.CanManage,
		u.CanAutoreg,
		u.ValidForever,
		u.ValidStart.Unix(),
		u.ValidEnd.Unix(),
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	u.ID = int(id)
	return nil
}

func (s *UserStore) Delete(u *models.User) error {
	if u.ID == 0 {
		return nil
	}

	sql := `DELETE FROM "user" WHERE "id" = ?`
	_, err := s.e.DB.Exec(sql, u.ID)
	return err
}
