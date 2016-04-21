package dhcp

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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

func (p *parser) parse() (*Config, error) {
	p.c = &Config{}

	for p.scan() {
		line := bytes.TrimSpace(p.s.Bytes())
		// Skip empty lines and comments
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		lineParts := bytes.SplitN(line, " ", 1)
		switch string(lineParts[0]) {
		case "global":
			gc, err := p.parseGlobal()
			if err != nil {
				return nil, err
			}
			p.c.Global = gc
		case "network":
			net, err := p.parseNetwork(lineParts[1])
			if err != nil {
				return nil, err
			}
			p.c.Networks = append(p.c.Networks, net)
		default:
			return nil, newError(p.line, "unexpected %s", lineParts[0])
		}
	}

	return config, p.s.Err()
}

func (p *parser) parseGlobal() (*GlobalConfig, error) {
	g := &Global{}
}

func (p *parser) parseSetting(s *Settings) error {
	return nil
}

func (p *parser) parseNetwork(header []byte) (*NetworkConfig, error) {
	return nil, nil
}

func (p *parser) parseSubnet(header []byte) (*SubnetConfig, error) {
	return nil, nil
}

func (p *parser) parsePool() (*PoolConfig, error) {
	return nil, nil
}
