// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

// ParseFile takes the file name to a configuration file.
// It will attempt to parse the file using the PG-DHCP configuration
// format. If an error occures config will be nil.
func ParseFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return newParser(bufio.NewScanner(file)).parse()
}

type parseError struct {
	line    int
	message string
}

func newError(line int, message string, v ...interface{}) parseError {
	return parseError{
		line:    line,
		message: fmt.Sprintf(message, v...),
	}
}

func (p parseError) Error() string {
	return fmt.Sprintf("Error %s on line %d", p.message, p.line)
}

type parser struct {
	c    *Config
	s    *bufio.Scanner
	line int
}

func newParser(s *bufio.Scanner) *parser {
	return &parser{s: s}
}

func (p *parser) scan() bool {
	if p.s.Scan() {
		p.line++
		return true
	}
	return false
}

func (p *parser) nextLine() []byte {
	for p.scan() {
		line := bytes.TrimSpace(p.s.Bytes())
		// Skip empty lines and comments
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		return line
	}
	return nil
}

func (p *parser) parse() (*Config, error) {
	p.c = newConfig()

	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.SplitN(line, []byte(" "), 2)
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[tkGlobal]) {
			if p.c.global != nil {
				return nil, newError(p.line, "unexpected 'global'")
			}
			if len(p.c.networks) > 0 {
				return nil, newError(p.line, "global settings must come before network settings")
			}
			gc, err := p.parseGlobal()
			if err != nil {
				return nil, err
			}
			p.c.global = gc
			continue
		} else if bytes.Equal(keyword, tokens[tkNetwork]) {
			if p.c.global == nil {
				p.c.global = newGlobal()
			}
			net, err := p.parseNetwork(lineParts[1])
			if err != nil {
				return nil, err
			}
			net.global = p.c.global
			p.c.networks[net.name] = net
			continue
		} else {
			return nil, newError(p.line, `unexpected "%s"`, keyword)
		}
	}

	return p.c, p.s.Err()
}

func (p *parser) parseGlobal() (*global, error) {
	g := newGlobal()
	// Modes: 0 Global, 1 Global registered, 2 Global unregistered
	mode := 0
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.SplitN(line, []byte(" "), 2)
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[tkEnd]) {
			if mode != 0 {
				mode = 0
				continue
			}
			break
		} else if bytes.Equal(keyword, tokens[tkServerIdentifier]) {
			ip := net.ParseIP(string(lineParts[1]))
			if ip == nil {
				return nil, newError(p.line, `invalid IP address '%s'`, lineParts[1])
			}
			g.serverIdentifier = ip
			continue
		} else if bytes.Equal(keyword, tokens[tkFreeLeaseAfter]) {
			time, err := strconv.Atoi(string(lineParts[1]))
			if err != nil {
				return nil, newError(p.line, "invalid value for free-lease-after")
			}

			if mode == 1 {
				g.registeredSettings.freeLeaseAfter = time
				continue
			} else if mode == 2 {
				g.unregisteredSettings.freeLeaseAfter = time
				continue
			}

			return nil, newError(p.line, "free-lease-after can only be in a global registered or unregistered block")
		} else if bytes.Equal(keyword, tokens[tkRegistered]) {
			mode = 1
			continue
		} else if bytes.Equal(keyword, tokens[tkUnregistered]) {
			mode = 2
			continue
		} else if isSetting(keyword) {
			var err error
			if mode == 1 {
				err = p.parseSetting(line, g.registeredSettings)
			} else if mode == 2 {
				err = p.parseSetting(line, g.unregisteredSettings)
			} else {
				err = p.parseSetting(line, g.settings)
			}
			if err != nil {
				return nil, err
			}
		} else {
			return nil, newError(p.line, `unexpected "%s"`, keyword)
		}
	}
	if g.settings.defaultLeaseTime == 0 {
		g.settings.defaultLeaseTime = time.Duration(12) * time.Hour // 12 hours
	}
	if g.settings.maxLeaseTime == 0 {
		g.settings.maxLeaseTime = time.Duration(12) * time.Hour // 12 hours
	}
	if g.unregisteredSettings.freeLeaseAfter == 0 {
		g.unregisteredSettings.freeLeaseAfter = 3600 // 1 hour
	}
	if g.registeredSettings.freeLeaseAfter == 0 {
		g.registeredSettings.freeLeaseAfter = 604800 // 1 week
	}
	return g, nil
}

func (p *parser) parseSetting(line []byte, s *settings) error {
	lineParts := bytes.SplitN(line, []byte(" "), 2)
	keyword := lineParts[0]

	if bytes.Equal(keyword, tokens[tkOption]) {
		lineParts := bytes.SplitN(lineParts[1], []byte(" "), 2)
		if len(lineParts) < 2 {
			return newError(p.line, `option %s has no value`, lineParts[1])
		}
		oc, ok := options[string(lineParts[0])]
		if !ok {
			return newError(p.line, `unknown option "%s"`, lineParts[0])
		}
		data, err := p.parseDHCPOption(oc, lineParts[1])
		if err != nil {
			return err
		}
		s.options[oc] = data
	} else if bytes.Equal(keyword, tokens[tkDefaultLeaseTime]) {
		i, err := strconv.Atoi(string(lineParts[1]))
		if err != nil {
			return newError(p.line, `invalid seconds amount: %s`, lineParts[1])
		}
		s.defaultLeaseTime = time.Duration(i) * time.Second
	} else if bytes.Equal(keyword, tokens[tkMaxLeaseTime]) {
		i, err := strconv.Atoi(string(lineParts[1]))
		if err != nil {
			return newError(p.line, `invalid seconds amount: %s`, lineParts[1])
		}
		s.maxLeaseTime = time.Duration(i) * time.Second
	} else {
		return newError(p.line, `unexpected "%s"`, keyword)
	}
	return nil
}

func (p *parser) parseDHCPOption(opcode dhcp4.OptionCode, data []byte) ([]byte, error) {
	switch opcode {
	case dhcp4.OptionSubnetMask:
		mask := net.ParseIP(string(data))
		if mask == nil {
			return nil, newError(p.line, `invalid subnet mask "%s"`, data)
		}
		return []byte(mask.To4()), nil
	case dhcp4.OptionDomainName:
		c := make([]byte, len(data))
		copy(c, data)
		return c, nil
	case dhcp4.OptionBroadcastAddress:
		broadcast := net.ParseIP(string(data))
		if broadcast == nil {
			return nil, newError(p.line, `invalid broadcast address "%s"`, data)
		}
		return []byte(broadcast.To4()), nil
	case dhcp4.OptionRouter:
		routersIn := bytes.Split(data, []byte(","))
		var routersOut []byte
		for _, r := range routersIn {
			router := net.ParseIP(string(bytes.TrimSpace(r)))
			if router == nil {
				return nil, newError(p.line, `invalid router address "%s"`, r)
			}
			routersOut = bytes.Join([][]byte{routersOut, router.To4()}, []byte(""))
		}
		return routersOut, nil
	case dhcp4.OptionDomainNameServer:
		serversIn := bytes.Split(data, []byte(","))
		var serversOut []byte
		for _, r := range serversIn {
			server := net.ParseIP(string(bytes.TrimSpace(r)))
			if server == nil {
				return nil, newError(p.line, `invalid DNS address "%s"`, r)
			}
			serversOut = bytes.Join([][]byte{serversOut, server.To4()}, []byte(""))
		}
		return serversOut, nil
	case dhcp4.OptionNetworkTimeProtocolServers:
		serversIn := bytes.Split(data, []byte(","))
		var serversOut []byte
		for _, r := range serversIn {
			server := net.ParseIP(string(bytes.TrimSpace(r)))
			if server == nil {
				return nil, newError(p.line, `invalid NTP address "%s"`, r)
			}
			serversOut = bytes.Join([][]byte{serversOut, server.To4()}, []byte(""))
		}
		return serversOut, nil
	}
	return data, nil
}

func (p *parser) parseSettingBlock() (*settings, error) {
	s := newSettingsBlock()
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		if bytes.Equal(line, tokens[tkEnd]) {
			break
		}

		if err := p.parseSetting(line, s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (p *parser) parseNetwork(header []byte) (*network, error) {
	if len(header) == 0 {
		return nil, newError(p.line, "network block requires a name")
	}
	n := newNetwork(string(header))
	// Mode 0 - network level, 1 - registered, 2 - unregistered
	mode := 0
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.SplitN(line, []byte(" "), 2)
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[tkEnd]) {
			if mode != 0 {
				mode = 0
				continue
			}
			break
		} else if isSetting(keyword) {
			var err error
			if mode == 1 {
				err = p.parseSetting(line, n.registeredSettings)
			} else if mode == 2 {
				err = p.parseSetting(line, n.unregisteredSettings)
			} else {
				err = p.parseSetting(line, n.settings)
			}
			if err != nil {
				return nil, err
			}
			continue
		} else if bytes.Equal(keyword, tokens[tkRegistered]) {
			mode = 1
			continue
		} else if bytes.Equal(keyword, tokens[tkUnregistered]) {
			mode = 2
			continue
		} else if bytes.Equal(keyword, tokens[tkSubnet]) {
			if mode == 0 {
				return nil, newError(p.line, "subnet must be in a registered or unregistered block")
			}
			if len(lineParts) < 2 {
				return nil, newError(p.line, "subnet block must have a CIDR formatted network")
			}
			sub, err := p.parseSubnet(lineParts[1], (mode == 2))
			if err != nil {
				return nil, err
			}
			sub.network = n
			n.subnets = append(n.subnets, sub)
		} else {
			return nil, newError(p.line, `unknown identifier %s`, keyword)
		}
	}
	return n, nil
}

func (p *parser) parseSubnet(header []byte, allowUnknown bool) (*subnet, error) {
	if len(header) == 0 {
		return nil, newError(p.line, "subnet block must have a CIDR formatted network")
	}
	s := newSubnet()
	_, sub, err := net.ParseCIDR(string(header))
	if err != nil {
		return nil, newError(p.line, `invalid CIDR formatted network: %s`, header)
	}
	s.net = sub
	s.allowUnknown = allowUnknown
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.Split(line, []byte(" "))
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[tkEnd]) {
			break
		} else if isSetting(keyword) {
			if err := p.parseSetting(line, s.settings); err != nil {
				return nil, err
			}
			continue
		} else if bytes.Equal(keyword, tokens[tkPool]) {
			pool, err := p.parsePool(nil, nil)
			if err != nil {
				return nil, err
			}
			pool.subnet = s
			s.pools = append(s.pools, pool)
			continue
		} else if bytes.Equal(keyword, tokens[tkRange]) {
			// Implict pool section. The range keyword start an implict pool
			// All subsequent lines will be consumed by the pool
			if len(lineParts) != 3 {
				return nil, newError(p.line, "range requires a start and an end")
			}
			pool, err := p.parsePool(lineParts[1], lineParts[2])
			if err != nil {
				return nil, err
			}
			pool.subnet = s
			s.pools = append(s.pools, pool)
			break
		} else {
			return nil, newError(p.line, `unknown keyword "%s"`, keyword)
		}
	}
	if _, ok := s.settings.options[dhcp4.OptionSubnetMask]; !ok {
		s.settings.options[dhcp4.OptionSubnetMask] = []byte(s.net.Mask)
	}
	return s, nil
}

func (p *parser) parsePool(rangeStart, rangeEnd []byte) (*pool, error) {
	pool := newPool()
	if rangeStart != nil {
		pool.rangeStart = net.ParseIP(string(rangeStart))
		if pool.rangeStart == nil {
			return nil, newError(p.line, `invalid range start "%s"`, rangeStart)
		}
	}
	if rangeEnd != nil {
		pool.rangeEnd = net.ParseIP(string(rangeEnd))
		if pool.rangeEnd == nil {
			return nil, newError(p.line, `invalid range start "%s"`, rangeEnd)
		}
	}
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.Split(line, []byte(" "))
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[tkEnd]) {
			break
		} else if isSetting(keyword) {
			if err := p.parseSetting(line, pool.settings); err != nil {
				return nil, err
			}
			continue
		} else if bytes.Equal(keyword, tokens[tkRange]) {
			if len(lineParts) != 3 {
				return nil, newError(p.line, "range requires a start and an end")
			}

			pool.rangeStart = net.ParseIP(string(lineParts[1]))
			if pool.rangeStart == nil {
				return nil, newError(p.line, `invalid range start "%s"`, lineParts[1])
			}

			pool.rangeEnd = net.ParseIP(string(lineParts[2]))
			if pool.rangeEnd == nil {
				return nil, newError(p.line, `invalid range start "%s"`, lineParts[2])
			}
		} else {
			return nil, newError(p.line, `unknown keyword "%s"`, keyword)
		}
	}
	return pool, nil
}
