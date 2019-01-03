// Original implementation: 2014 Skagerrak Software - http://www.skagerraksoftware.com/
// Modifications: 2017 Lee Keitel

package dhcp4

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1500)
		},
	}
)

// A Handler takes a DHCP request packet and generates a response to the client
type Handler interface {
	ServeDHCP(req Packet, msgType MessageType, options Options) Packet
}

// ServeConn is the bare minimum connection functions required by Serve()
// It allows you to create custom connections for greater control,
// such as ServeIfConn (see serverif.go), which locks to a given interface.
type ServeConn interface {
	ReadFrom(b []byte) (n int, addr net.Addr, err error)
	WriteTo(b []byte, addr net.Addr) (n int, err error)
}

// Serve takes a ServeConn (such as a net.PacketConn) that it uses for both
// reading and writing DHCP packets. Every packet is passed to the handler,
// which processes it and optionally return a response packet for writing back
// to the network.
//
// To capture limited broadcast packets (sent to 255.255.255.255), you must
// listen on a socket bound to IP_ADDRANY (0.0.0.0). This means that broadcast
// packets sent to any interface on the system may be delivered to this
// socket.  See: https://code.google.com/p/go/issues/detail?id=7106
//
// Additionally, response packets may not return to the same
// interface that the request was received from.  Writing a custom ServeConn,
// can provide a workaround to this problem.
func Serve(conn ServeConn, handler Handler, workers int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			return
		}
	}()

	taskQueue := startWorkers(workers, conn, handler)

	for {
		buffer := bufferPool.Get().([]byte)
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			close(taskQueue)
			return err
		}
		if n < 240 { // Packet too small to be DHCP
			continue
		}
		req := Packet(buffer[:n])
		if req.HLen() > 16 { // Invalid size
			continue
		}

		select {
		case taskQueue <- job{p: req, from: addr}:
		default:
			fmt.Println("Task queue full")
			bufferPool.Put(buffer)
		}
	}
}

func process(conn ServeConn, p Packet, handler Handler, from net.Addr) {
	options := p.ParseOptions()

	t := options[OptionDHCPMessageType]
	if len(t) != 1 {
		return
	}

	reqType := MessageType(t[0])
	if reqType < Discover || reqType > Inform {
		return
	}

	if res := handler.ServeDHCP(p, reqType, options); res != nil {
		// If coming from a relay, unicast back
		if !p.GIAddr().Equal(net.IPv4zero) {
			if _, e := conn.WriteTo(res, from); e != nil {
				panic(e)
			}
			return
		}

		ipStr, portStr, err := net.SplitHostPort(from.String())
		if err != nil {
			return
		}

		// If IP not available or broadcast bit is set, broadcast
		if net.ParseIP(ipStr).Equal(net.IPv4zero) || p.Broadcast() {
			port, _ := strconv.Atoi(portStr)
			from = &net.UDPAddr{IP: net.IPv4bcast, Port: port}
		}
		if _, e := conn.WriteTo(res, from); e != nil {
			panic(e)
		}
	}
}

type job struct {
	p    Packet
	from net.Addr
}

func startWorkers(num int, conn ServeConn, handler Handler) chan job {
	tasks := make(chan job, num*2)

	for i := 1; i <= num; i++ {
		fmt.Printf("Starting worker %d\n", i)
		go worker(conn, handler, tasks)
	}

	return tasks
}

func worker(conn ServeConn, handler Handler, tasks <-chan job) {
	for j := range tasks {
		process(conn, j.p, handler, j.from)
		bufferPool.Put([]byte(j.p))
	}
	fmt.Println("Working stopping")
}
