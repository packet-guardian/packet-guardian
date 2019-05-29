package models

import "net"

type TestUserStore struct{}

func (s *TestUserStore) Save(u *User) error                   { return nil }
func (s *TestUserStore) Delete(u *User) error                 { return nil }
func (s *TestUserStore) GetPassword(u string) (string, error) { return "", nil }

type TestDeviceStore struct{}

func (s *TestDeviceStore) Save(d *Device) error                 { return nil }
func (s *TestDeviceStore) Delete(d *Device) error               { return nil }
func (s *TestDeviceStore) DeleteAllDeviceForUser(u *User) error { return nil }

type TestLeaseStore struct{}

func (s *TestLeaseStore) GetLeaseHistory(m net.HardwareAddr) ([]LeaseHistory, error) { return nil, nil }
func (s *TestLeaseStore) GetLatestLease(m net.HardwareAddr) LeaseHistory             { return nil }

type TestBlacklistItem struct{}

func (i *TestBlacklistItem) Blacklist()                  {}
func (i *TestBlacklistItem) Unblacklist()                {}
func (i *TestBlacklistItem) IsBlacklisted(s string) bool { return false }
func (i *TestBlacklistItem) Save(s string) error         { return nil }
