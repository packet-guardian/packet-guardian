// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// ParseFile takes the file name to a configuration file.
// It will attempt to parse the file using the PG-DHCP configuration
// format. If an error occures config will be nil.
func ParseFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return newParser(bufio.NewReader(file)).parse()
}

type parser struct {
	l *lexer
	c *Config
}

func newParser(r *bufio.Reader) *parser {
	return &parser{l: newLexer(r)}
}

func (p *parser) parse() (*Config, error) {
	p.c = newConfig()

mainLoop:
	for {
		tok := p.l.next()
		var err error
		switch tok.token {
		case COMMENT, EOL:
			continue
		case GLOBAL:
			err = p.parseGlobal()
		case NETWORK:
			err = p.parseNetwork()
		case INCLUDE:
			n := p.l.next()
			if n.token != STRING {
				return nil, fmt.Errorf("Include must be a file path on line %d", n.line)
			}
			file, err := os.Open(n.value.(string))
			if err != nil {
				return nil, fmt.Errorf("Error including file %v: %v", n.value, err)
			}
			p.l.pushReader(bufio.NewReader(file))
			continue
		case EOF:
			break mainLoop
		default:
			return nil, fmt.Errorf("Invalid token on line %d: %s", tok.line, tok.string())
		}
		if err != nil {
			return nil, err
		}
	}

	for _, n := range p.c.networks {
		n.global = p.c.global
	}
	return p.c, nil
}

func (p *parser) parseGlobal() error {
mainLoop:
	for {
		tok := p.l.next()
		switch tok.token {
		case COMMENT, EOL:
			continue
		case EOF, END:
			break mainLoop
		case SERVER_IDENTIFIER:
			addr := p.l.next()
			if addr.token != IP_ADDRESS {
				return fmt.Errorf("Expected IP address on line %d", addr.line)
			}
			p.c.global.serverIdentifier = addr.value.(net.IP)
		case REGISTERED:
			s, err := p.parseSettingsBlock()
			if err != nil {
				return err
			}
			p.c.global.registeredSettings = s
			p.l.next() // Consume END from block
		case UNREGISTERED:
			s, err := p.parseSettingsBlock()
			if err != nil {
				return err
			}
			p.c.global.unregisteredSettings = s
			p.l.next() // Consume END from block
		default:
			if tok.token.isSetting() {
				p.l.unread()
				err := p.parseSetting(p.c.global.settings)
				if err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("Unexpected token %s on line %d in global", tok.string(), tok.line)
		}
	}
	return nil
}

func (p *parser) parseNetwork() error {
	nameToken := p.l.next()
	if nameToken.token != STRING {
		return fmt.Errorf("Expected STRING on line %d", nameToken.line)
	}
	name := strings.ToLower(nameToken.value.(string))

	if _, exists := p.c.networks[name]; exists {
		return fmt.Errorf("Network %s already declared, line %d", name, nameToken.line)
	}
	netBlock := newNetwork(name)
	mode := 0 // 0 = root, 1 = registered, 2 = unregistered
mainLoop:
	for {
		tok := p.l.next()
		switch tok.token {
		case EOF:
			break mainLoop
		case COMMENT, EOL:
			continue
		case SUBNET:
			shortSyntax := false
			if mode == 0 {
				mode = 2
				shortSyntax = true
			}
			subnet, err := p.parseSubnet()
			if err != nil {
				return err
			}
			if mode == 2 { // Unregistered block
				subnet.allowUnknown = true
			}
			subnet.network = netBlock
			netBlock.subnets = append(netBlock.subnets, subnet)
			if shortSyntax {
				mode = 0
			}
		case REGISTERED:
			if mode == 0 {
				mode = 1
				continue
			}
			return fmt.Errorf("Registered block not allowed on line %d", tok.line)
		case UNREGISTERED:
			if mode == 0 {
				mode = 2
				continue
			}
			return fmt.Errorf("Unregistered block not allowed on line %d", tok.line)
		case END:
			if mode == 0 { // Exit from root network block
				break mainLoop
			} else { // Exit from un/registered block
				mode = 0
			}
		default:
			if tok.token.isSetting() {
				p.l.unread()
				block := netBlock.settings
				if mode == 1 {
					block = netBlock.registeredSettings
				} else if mode == 2 {
					block = netBlock.unregisteredSettings
				}
				err := p.parseSetting(block)
				if err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("Unexpected token %s on line %d in network", tok.string(), tok.line)
		}
	}
	p.c.networks[name] = netBlock
	return nil
}

func (p *parser) parseSubnet() (*subnet, error) {
	ipAddr := p.l.next()
	if ipAddr.token != IP_ADDRESS {
		return nil, fmt.Errorf("Expected IP address on line %d", ipAddr.line)
	}

	netmask := p.l.next()
	if netmask.token != IP_ADDRESS {
		return nil, fmt.Errorf("Expected IP address on line %d", netmask.line)
	}
	sub := newSubnet()
	sub.net = &net.IPNet{
		IP:   ipAddr.value.(net.IP),
		Mask: net.IPMask(netmask.value.(net.IP)),
	}

mainLoop:
	for {
		tok := p.l.next()
		switch tok.token {
		case COMMENT, EOL:
			continue
		case EOF, END:
			break mainLoop
		case POOL:
			subPool, err := p.parsePool()
			if err != nil {
				return nil, err
			}
			subPool.subnet = sub
			sub.pools = append(sub.pools, subPool)
		case RANGE:
			p.l.unread()
			subPool, err := p.parsePool() // Start with range statement
			if err != nil {
				return nil, err
			}
			subPool.subnet = sub
			sub.pools = append(sub.pools, subPool)
			p.l.unread() // Reread END token
		default:
			if tok.token.isSetting() {
				p.l.unread()
				err := p.parseSetting(sub.settings)
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("Unexpected token %s on line %d in subnet", tok.string(), tok.line)
		}
	}
	return sub, nil
}

func (p *parser) parsePool() (*pool, error) {
	nPool := newPool()

mainLoop:
	for {
		tok := p.l.next()
		switch tok.token {
		case COMMENT, EOL:
			continue
		case EOF, END:
			break mainLoop
		case RANGE:
			if nPool.rangeStart != nil {
				// If we encounter another range statement, assume it's a new Pool block
				p.l.unread()
				break mainLoop
			}
			startIP := p.l.next()
			if startIP.token != IP_ADDRESS {
				return nil, fmt.Errorf("Expected IP address on line %d, got %s", startIP.line, startIP.string())
			}
			nPool.rangeStart = startIP.value.(net.IP)

			endIP := p.l.next()
			if endIP.token != IP_ADDRESS {
				return nil, fmt.Errorf("Expected IP address on line %d, got %s", endIP.line, endIP.string())
			}
			nPool.rangeEnd = endIP.value.(net.IP)
		default:
			if tok.token.isSetting() {
				p.l.unread()
				err := p.parseSetting(nPool.settings)
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("Unexpected token %s on line %d in pool", tok.string(), tok.line)
		}
	}
	return nPool, nil
}

func (p *parser) parseSettingsBlock() (*settings, error) {
	s := newSettingsBlock()

	for {
		tok := p.l.next()
		if tok.token == COMMENT || tok.token == EOL {
			continue
		}
		p.l.unread()
		if !tok.token.isSetting() {
			break
		}
		err := p.parseSetting(s)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (p *parser) parseSetting(setBlock *settings) error {
	tok := p.l.next()

	switch tok.token {
	case COMMENT, EOF:
		return nil
	case OPTION:
		p.parseOption()
		return nil
	case DEFAULT_LEASE_TIME:
		tokn := p.l.next()
		if tokn.token != NUMBER {
			return fmt.Errorf("Expected number on line %d", tokn.line)
		}
		setBlock.defaultLeaseTime = time.Duration(tokn.value.(uint64)) * time.Second
		return nil
	case MAX_LEASE_TIME:
		tokn := p.l.next()
		if tokn.token != NUMBER {
			return fmt.Errorf("Expected number on line %d", tokn.line)
		}
		setBlock.maxLeaseTime = time.Duration(tokn.value.(uint64)) * time.Second
		return nil
	case FREE_LEASE_AFTER:
		tokn := p.l.next()
		if tokn.token != NUMBER {
			return fmt.Errorf("Expected number on line %d", tokn.line)
		}
		setBlock.freeLeaseAfter = time.Duration(tokn.value.(uint64)) * time.Second
		return nil
	}

	return fmt.Errorf("Unexpected token %s on line %d in settings", tok.string(), tok.line)
}

func (p *parser) parseOption() { p.l.untilNext(EOL) }
