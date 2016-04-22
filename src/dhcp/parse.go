package dhcp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

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
	p.c = &Config{}

	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.SplitN(line, []byte(" "), 2)
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[TkGlobal]) {
			if p.c.Global != nil {
				return nil, newError(p.line, "unexpected 'global'")
			}
			gc, err := p.parseGlobal()
			if err != nil {
				return nil, err
			}
			p.c.Global = gc
			continue
		} else if bytes.Equal(keyword, tokens[TkNetwork]) {
			net, err := p.parseNetwork(lineParts[1])
			if err != nil {
				return nil, err
			}
			p.c.Networks = append(p.c.Networks, net)
			continue
		} else {
			return nil, newError(p.line, `unexpected "%s"`, keyword)
		}
	}

	return p.c, p.s.Err()
}

func (p *parser) parseGlobal() (*Global, error) {
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

		if bytes.Equal(keyword, tokens[TkEnd]) {
			if mode != 0 {
				mode = 0
				continue
			}
			break
		} else if bytes.Equal(keyword, tokens[TkServerIdentifier]) {
			ip := net.ParseIP(string(lineParts[1]))
			if ip == nil {
				return nil, newError(p.line, `invalid IP address '%s'`, lineParts[1])
			}
			g.ServerIdentifier = ip
			continue
		} else if bytes.Equal(keyword, tokens[TkRegistered]) {
			mode = 1
			continue
		} else if bytes.Equal(keyword, tokens[TkUnregistered]) {
			mode = 2
			continue
		} else if isSetting(keyword) {
			var err error
			if mode == 1 {
				err = p.parseSetting(line, g.RegisteredSettings)
			} else if mode == 2 {
				err = p.parseSetting(line, g.UnregisteredSettings)
			} else {
				err = p.parseSetting(line, g.Settings)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	return g, nil
}

func (p *parser) parseSetting(line []byte, s *Settings) error {
	lineParts := bytes.SplitN(line, []byte(" "), 2)
	keyword := lineParts[0]

	if bytes.Equal(keyword, tokens[TkOption]) {
		lineParts := bytes.SplitN(lineParts[1], []byte(" "), 2)
		if len(lineParts) < 2 {
			return newError(p.line, `option %s has no value`, lineParts[1])
		}
		oc, ok := options[string(lineParts[0])]
		if !ok {
			return newError(p.line, `unknown option "%s"`, lineParts[0])
		}
		s.Options[oc] = lineParts[1]
	} else if bytes.Equal(keyword, tokens[TkDefaultLeaseTime]) {
		i, err := strconv.Atoi(string(lineParts[1]))
		if err != nil {
			return newError(p.line, `invalid seconds amount: %s`, lineParts[1])
		}
		s.DefaultLeaseTime = time.Duration(i) * time.Second
	} else if bytes.Equal(keyword, tokens[TkMaxLeaseTime]) {
		i, err := strconv.Atoi(string(lineParts[1]))
		if err != nil {
			return newError(p.line, `invalid seconds amount: %s`, lineParts[1])
		}
		s.MaxLeaseTime = time.Duration(i) * time.Second
	} else {
		return newError(p.line, `unexpected "%s"`, keyword)
	}
	return nil
}

func (p *parser) parseSettingBlock() (*Settings, error) {
	s := newSettingsBlock()
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		if bytes.Equal(line, tokens[TkEnd]) {
			break
		}

		if err := p.parseSetting(line, s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (p *parser) parseNetwork(header []byte) (*Network, error) {
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

		if bytes.Equal(keyword, tokens[TkEnd]) {
			if mode != 0 {
				mode = 0
				continue
			}
			break
		} else if isSetting(keyword) {
			var err error
			if mode == 1 {
				err = p.parseSetting(line, n.RegisteredSettings)
			} else if mode == 2 {
				err = p.parseSetting(line, n.UnregisteredSettings)
			} else {
				err = p.parseSetting(line, n.Settings)
			}
			if err != nil {
				return nil, err
			}
			continue
		} else if bytes.Equal(keyword, tokens[TkRegistered]) {
			mode = 1
			continue
		} else if bytes.Equal(keyword, tokens[TkUnregistered]) {
			mode = 2
			continue
		} else if bytes.Equal(keyword, tokens[TkSubnet]) {
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
			n.Subnets = append(n.Subnets, sub)
		} else {
			return nil, newError(p.line, `unknown identifier %s`, keyword)
		}
	}
	return n, nil
}

func (p *parser) parseSubnet(header []byte, allowUnknown bool) (*Subnet, error) {
	if len(header) == 0 {
		return nil, newError(p.line, "subnet block must have a CIDR formatted network")
	}
	s := newSubnet()
	_, sub, err := net.ParseCIDR(string(header))
	if err != nil {
		return nil, newError(p.line, `invalid CIDR formatted network: %s`, header)
	}
	s.Network = sub
	s.AllowUnknown = allowUnknown
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.SplitN(line, []byte(" "), 2)
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[TkEnd]) {
			break
		}
		// TODO: Actually parse the subnet piece
	}
	return s, nil
}

func (p *parser) parsePool() (*Pool, error) {
	pool := newPool()
	pool.RangeStart = net.IP{}
	pool.RangeEnd = net.IP{}
	for {
		line := p.nextLine()
		if line == nil {
			break
		}

		lineParts := bytes.SplitN(line, []byte(" "), 2)
		keyword := lineParts[0]

		if bytes.Equal(keyword, tokens[TkEnd]) {
			break
		}
		// TODO: Actually parse the pool piece
	}
	return pool, nil
	return nil, nil
}
