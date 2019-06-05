// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stores

import (
	"errors"
	"strings"
	"time"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
)

var appUserStore UserStore

type UserStore interface {
	GetUserByUsername(username string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	SearchUsersByField(field, pattern string) ([]*models.User, error)
	GetPassword(username string) (string, error)
	Save(u *models.User) error
	Delete(u *models.User) error
}

type userStore struct {
	e *common.Environment
}

func newUserStore(e *common.Environment) *userStore {
	return &userStore{
		e: e,
	}
}

func GetUserStore(e *common.Environment) UserStore {
	if appUserStore == nil {
		appUserStore = newUserStore(e)
	}
	return appUserStore
}

func (s *userStore) GetUserByUsername(username string) (*models.User, error) {
	if username == "" {
		return models.NewUser(s.e, s, NewBlacklistItem(GetBlacklistStore(s.e)), ""), nil
	}

	username = strings.ToLower(username)

	sql := `WHERE "username" = ?`
	users, err := s.getUsersFromDatabase(sql, username)
	if len(users) == 0 {
		u := models.NewUser(s.e, s, NewBlacklistItem(GetBlacklistStore(s.e)), username)
		return u, err
	}
	return users[0], nil
}

func (s *userStore) GetAllUsers() ([]*models.User, error) {
	sql := `ORDER BY "username"`
	if s.e.DB.Driver == "sqlite" {
		sql += " COLLATE NOCASE"
	}
	sql += " ASC"
	return s.getUsersFromDatabase(sql)
}

func (s *userStore) SearchUsersByField(field, pattern string) ([]*models.User, error) {
	sql := `WHERE "` + field + `" LIKE ?`
	return s.getUsersFromDatabase(sql, pattern)
}

func (s *userStore) getUsersFromDatabase(where string, values ...interface{}) ([]*models.User, error) {
	sql := `SELECT "id", "username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "can_autoreg", "valid_forever", "valid_start", "valid_end", "ui_group", "api_group", "allow_status_api" FROM "user" ` + where
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
		var uiGroup string
		var apiGroup string
		var allowStatusAPI bool

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
			&uiGroup,
			&apiGroup,
			&allowStatusAPI,
		)
		if err != nil {
			continue
		}

		user := models.NewUser(s.e, s, NewBlacklistItem(GetBlacklistStore(s.e)), username)
		user.ID = id
		user.HasPassword = (password != "")
		user.DeviceLimit = models.UserDeviceLimit(deviceLimit)
		user.ValidStart = time.Unix(validStart, 0)
		user.ValidEnd = time.Unix(validEnd, 0)
		user.ValidForever = validForever
		user.CanManage = canManage
		user.CanAutoreg = canAutoreg
		user.Rights = models.ViewOwn
		user.UIGroup = uiGroup
		user.APIGroup = apiGroup
		user.AllowStatusAPI = allowStatusAPI

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
		user.LoadRights() // Above, all rights are overriden so we need to reapply admin and configured rights
		results = append(results, user)
	}
	return results, nil
}

func (s *userStore) GetPassword(username string) (string, error) {
	result := s.e.DB.QueryRow(`SELECT "password" FROM "user" WHERE "username" = ?`, username)
	var p string
	err := result.Scan(&p)
	return p, err
}

func (s *userStore) Save(u *models.User) error {
	if u.ID == 0 {
		return s.saveNew(u)
	}
	return s.updateExisting(u)
}

func (s *userStore) updateExisting(u *models.User) error {
	sql := `UPDATE "user" SET "device_limit"=?, "default_expiration"=?, "expiration_type"=?, "can_manage"=?, "can_autoreg"=?, "valid_forever"=?, "valid_start"=?, "valid_end"=?, "ui_group"=?, "api_group"=?, "allow_status_api"=?`

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
			u.UIGroup,
			u.APIGroup,
			u.AllowStatusAPI,
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
			u.UIGroup,
			u.APIGroup,
			u.AllowStatusAPI,
			u.ID,
		)
	}
	return err
}

func (s *userStore) saveNew(u *models.User) error {
	if u.Username == "" {
		return errors.New("Username cannot be empty")
	}

	sql := `INSERT INTO "user" ("username", "password", "device_limit", "default_expiration", "expiration_type", "can_manage", "can_autoreg", "valid_forever", "valid_start", "valid_end", "ui_group", "api_group", "allow_status_api") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`

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
		u.UIGroup,
		u.APIGroup,
		u.AllowStatusAPI,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	u.ID = int(id)
	return nil
}

func (s *userStore) Delete(u *models.User) error {
	if u.ID == 0 {
		return nil
	}

	sql := `DELETE FROM "user" WHERE "id" = ?`
	_, err := s.e.DB.Exec(sql, u.ID)
	return err
}
