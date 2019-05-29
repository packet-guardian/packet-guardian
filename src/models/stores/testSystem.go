package stores

import (
	"bytes"
	"net"

	"github.com/packet-guardian/dhcp-lib"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models"
)

type TestLeaseStore struct {
	Leases []*dhcp.Lease
}

func (s *TestLeaseStore) GetAllLeases() ([]*dhcp.Lease, error) {
	return s.Leases, nil
}
func (s *TestLeaseStore) GetLeaseByIP(ip net.IP) (*dhcp.Lease, error) {
	for _, l := range s.Leases {
		if l.IP.Equal(ip) {
			return l, nil
		}
	}
	return nil, nil
}
func (s *TestLeaseStore) GetRecentLeaseByMAC(mac net.HardwareAddr) (*dhcp.Lease, error) {
	for _, l := range s.Leases {
		if bytes.Equal(l.MAC, mac) {
			return l, nil
		}
	}
	return nil, nil
}
func (s *TestLeaseStore) GetAllLeasesByMAC(mac net.HardwareAddr) ([]*dhcp.Lease, error) {
	leases := make([]*dhcp.Lease, 0, 5)
	for _, l := range s.Leases {
		if bytes.Equal(l.MAC, mac) {
			leases = append(leases, l)
		}
	}
	return leases, nil
}
func (s *TestLeaseStore) CreateLease(lease *dhcp.Lease) error {
	s.Leases = append(s.Leases, lease)
	return nil
}
func (s *TestLeaseStore) GetLeaseHistory(mac net.HardwareAddr) ([]models.LeaseHistory, error) {
	return nil, nil
}
func (s *TestLeaseStore) UpdateLease(lease *dhcp.Lease) error { return nil }
func (s *TestLeaseStore) DeleteLease(lease *dhcp.Lease) error { return nil }
func (s *TestLeaseStore) SearchLeases(where string, vals ...interface{}) ([]*dhcp.Lease, error) {
	return nil, nil
}
func (s *TestLeaseStore) GetLatestLease(mac net.HardwareAddr) models.LeaseHistory { return nil }

type TestUserStore struct {
	Users []*models.User
}

func (s *TestUserStore) GetUserByUsername(username string) (*models.User, error) {
	for _, u := range s.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return models.NewUser(nil, s, NewBlacklistItem(&TestBlacklistStore{}), username), nil
}

func (s *TestUserStore) GetAllUsers() ([]*models.User, error) {
	return s.Users, nil
}

func (s *TestUserStore) SearchUsersByField(field, pattern string) ([]*models.User, error) {
	return nil, nil
}

func (s *TestUserStore) GetPassword(username string) (string, error) {
	for _, u := range s.Users {
		if u.Username == username {
			return u.Password, nil
		}
	}
	return "", nil
}
func (s *TestUserStore) Save(u *models.User) error   { return nil }
func (s *TestUserStore) Delete(u *models.User) error { return nil }

type TestDeviceStore struct {
	Devices []*models.Device
}

func (s *TestDeviceStore) GetDeviceByMAC(mac net.HardwareAddr) (*models.Device, error) {
	for _, d := range s.Devices {
		if bytes.Equal(d.MAC, mac) {
			return d, nil
		}
	}

	d := models.NewDevice(nil, s, &TestLeaseStore{}, NewBlacklistItem(&TestBlacklistStore{}))
	d.MAC = mac
	return d, nil
}

func (s *TestDeviceStore) GetDeviceByID(id int) (*models.Device, error) {
	for _, d := range s.Devices {
		if d.ID == id {
			return d, nil
		}
	}

	d := models.NewDevice(nil, s, &TestLeaseStore{}, NewBlacklistItem(&TestBlacklistStore{}))
	return d, nil
}
func (s *TestDeviceStore) GetFlaggedDevices() ([]*models.Device, error) {
	devices := make([]*models.Device, 0, 5)
	for _, d := range s.Devices {
		if d.Flagged {
			devices = append(devices, d)
		}
	}
	return devices, nil
}
func (s *TestDeviceStore) GetDevicesForUser(u *models.User) ([]*models.Device, error) {
	devices := make([]*models.Device, 0, 5)
	for _, d := range s.Devices {
		if d.Username == u.Username {
			devices = append(devices, d)
		}
	}
	return devices, nil
}
func (s *TestDeviceStore) GetDevicesForUserPage(u *models.User, page int) ([]*models.Device, error) {
	devices := make([]*models.Device, 0, 5)
	if page == 0 {
		for _, d := range s.Devices {
			if d.Username == u.Username {
				devices = append(devices, d)
			}
		}
	}
	return devices, nil
}
func (s *TestDeviceStore) GetDeviceCountForUser(u *models.User) (int, error) {
	devices := make([]*models.Device, 0, 5)
	for _, d := range s.Devices {
		if d.Username == u.Username {
			devices = append(devices, d)
		}
	}
	return len(devices), nil
}
func (s *TestDeviceStore) GetAllDevices(e *common.Environment) ([]*models.Device, error) {
	return s.Devices, nil
}
func (s *TestDeviceStore) SearchDevicesByField(field, pattern string) ([]*models.Device, error) {
	return nil, nil
}
func (s *TestDeviceStore) Save(d *models.Device) error                 { return nil }
func (s *TestDeviceStore) Delete(d *models.Device) error               { return nil }
func (s *TestDeviceStore) DeleteAllDeviceForUser(u *models.User) error { return nil }

type TestBlacklistStore struct {
	items map[string]bool
}

func (s *TestBlacklistStore) IsBlacklisted(key string) bool { return s.items[key] }
func (s *TestBlacklistStore) AddToBlacklist(key string) error {
	s.items[key] = true
	return nil
}
func (s *TestBlacklistStore) RemoveFromBlacklist(key string) error {
	delete(s.items, key)
	return nil
}
