package stores

type StoreCollection struct {
	Blacklist BlacklistStore
	Devices   DeviceStore
	Leases    LeaseStore
	Users     UserStore
}
