package dhcp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type lease struct {
	start time.Time
	end   time.Time
	mac   string
}

const leaseTimeLayout = "2006/01/02 15:04:05"

var leaseMutex = &sync.Mutex{}

func getLeaseFromFile(ip net.IP, leaseFile string) (*lease, error) {
	leaseMutex.Lock()
	defer leaseMutex.Unlock()
	f, err := os.Open(leaseFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return getLeaseReader(ip, f)
}

func getLeaseReader(ip net.IP, lf io.Reader) (*lease, error) {
	searchStr := []byte("lease " + ip.String() + " {")
	var latestLease *lease
	scanner := bufio.NewScanner(lf)
	for scanner.Scan() {
		if !bytes.Contains(scanner.Bytes(), searchStr) {
			continue
		}
		l := getLeaseInfo(scanner)
		if l != nil {
			latestLease = l
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if latestLease == nil {
		return nil, errors.New("No lease found")
	}
	return latestLease, nil
}

func getLeaseInfo(s *bufio.Scanner) *lease {
	l := &lease{}
	for s.Scan() {
		if strings.TrimSpace(s.Text()) == "}" {
			break
		}
		t := strings.TrimSpace(s.Text())
		t = t[:len(t)-1]
		line := strings.SplitN(t, " ", 3)
		if len(line) < 3 {
			continue
		}
		switch line[0] {
		case "starts":
			l.start, _ = time.Parse(leaseTimeLayout, line[2])
		case "ends":
			l.end, _ = time.Parse(leaseTimeLayout, line[2])
		case "hardware":
			l.mac = line[2]
		}
	}

	if l.start.IsZero() || l.end.IsZero() || l.mac == "" {
		return nil
	}
	return l
}
