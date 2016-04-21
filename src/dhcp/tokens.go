package dhcp

var Token int

const (
	TkEnd Token = iota
    TkGlobal
	TkNetwork
	TkSubnet
	TkPool
	TkRegistered
	TkUnregistered

	TkOption
	TkDefaultLeaseTime
	TkMaxLeaseTime
	TkServerIdentifier
    TkRange
)

var tokens = [...][]byte{
    TkEnd: []byte("end")
    TkGlobal: []byte("global")
	TkNetwork: []byte("network")
	TkSubnet: []byte("subnet")
	TkPool: []byte("pool")
	TkRegistered: []byte("registered")
	TkUnregistered: []byte("unregistered")

	TkOption: []byte("option")
	TkDefaultLeaseTime: []byte("default-lease-time")
	TkMaxLeaseTime: []byte("max-lease-time")
	TkServerIdentifier: []byte("server-identifier")
    TkRange: []byte("range")
}
