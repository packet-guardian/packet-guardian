// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"unicode"
)

type lexer struct {
	line     int
	r        *bufio.Reader
	buffer   []*lexToken
	prev     *lexToken
	readPrev bool
	readers  []*bufio.Reader
}

func newLexer(r *bufio.Reader) *lexer {
	return &lexer{
		r:       r,
		line:    1,
		readers: make([]*bufio.Reader, 0),
	}
}

func (l *lexer) pushReader(r *bufio.Reader) {
	l.readers = append(l.readers, l.r)
	l.r = r
}

func (l *lexer) popReader() bool {
	if len(l.readers) == 0 {
		return false
	}
	l.r = l.readers[len(l.readers)-1]
	l.readers = l.readers[0 : len(l.readers)-1]
	return true
}

// This function will make the lexer reread the previous token. This can
// only be used to reread one token.
func (l *lexer) unread() {
	l.readPrev = true
}

func (l *lexer) all() []*lexToken {
	var tokens []*lexToken
	for {
		tok := l.next()
		if tok.token == EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (l *lexer) untilNext(t token) []*lexToken {
	var tokens []*lexToken
	for {
		tok := l.next()
		if tok.token == t || tok.token == EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (l *lexer) next() *lexToken {
	if l.readPrev {
		l.readPrev = false
		return l.prev
	}

	if len(l.buffer) > 0 {
		tok := l.buffer[0]
		l.buffer = l.buffer[1:]
		l.prev = tok
		return tok
	}

	var tok []*lexToken // Some consumes produce multiple tokens

	for {
		c, err := l.r.ReadByte()
		if err != nil {
			if l.popReader() {
				continue
			}
			return &lexToken{token: EOF}
		}

		if c == '"' {
			tok = l.consumeString() // Start after double quote
			break
		} else if isNumber(c) || c == '-' {
			l.r.UnreadByte()
			tok = l.consumeNumeric()
			break
		} else if c == '\n' {
			l.line++
			tok = []*lexToken{&lexToken{token: EOL}}
			break
		} else if c == '#' {
			line := l.consumeLine()
			tok = []*lexToken{
				&lexToken{
					token: COMMENT,
					value: string(line),
				},
			}
			break
		} else if isLetter(c) {
			l.r.UnreadByte()
			tok = l.consumeIdent()
			break
		}
	}

	// Ensure all produced tokens have a line number
	for _, t := range tok {
		t.line = l.line
	}

	// This function only returns one token, if more were created,
	// add them to a buffer to be returned later
	if len(tok) > 1 {
		l.buffer = tok[1:]
	}

	l.prev = tok[0]
	return tok[0]
}

func (l *lexer) consumeString() []*lexToken {
	buf := bytes.Buffer{}
	for {
		b, err := l.r.ReadByte()
		if err != nil {
			return nil
		}
		if b == '"' {
			break
		}
		buf.WriteByte(b)
	}
	return []*lexToken{&lexToken{token: STRING, value: buf.String()}}
}

func (l *lexer) consumeLine() []byte {
	buf := bytes.Buffer{}
	for {
		b, err := l.r.ReadByte()
		if err != nil {
			return nil
		}
		if b == '\n' {
			l.r.UnreadByte()
			break
		}
		buf.WriteByte(b)
	}
	return buf.Bytes()
}

func (l *lexer) consumeNumeric() []*lexToken {
	buf := bytes.Buffer{}
	dotCount := 0
	hasSlash := false
	negative := false

	for {
		b, err := l.r.ReadByte()
		if err != nil {
			return nil
		}
		if isNumber(b) {
			buf.WriteByte(b)
			continue
		} else if b == '.' {
			buf.WriteByte(b)
			dotCount++
			continue
		} else if b == '/' {
			buf.WriteByte(b)
			hasSlash = true
			continue
		} else if b == '-' {
			negative = true
			continue
		}
		l.r.UnreadByte()
		break
	}

	toks := make([]*lexToken, 1)
	toks[0] = &lexToken{}
	if hasSlash && dotCount == 3 { // CIDR notation
		ip, network, err := net.ParseCIDR(buf.String())
		if err != nil {
			toks[0].token = ILLEGAL
		} else {
			toks[0].token = IP_ADDRESS
			toks[0].value = ip
			t := &lexToken{
				token: IP_ADDRESS,
				value: net.IP(network.Mask),
			}
			toks = append(toks, t)
		}
	} else if dotCount == 3 { // IP Address
		ip := net.ParseIP(buf.String())
		if ip == nil {
			toks[0].token = ILLEGAL
		} else {
			toks[0].token = IP_ADDRESS
			toks[0].value = ip
		}
	} else if dotCount == 0 { // Number
		if negative {
			num, err := strconv.ParseInt(buf.String(), 10, 64)
			if err != nil {
				toks[0].token = ILLEGAL
			} else {
				toks[0].token = NUMBER
				toks[0].value = num * -1
			}
		} else {
			num, err := strconv.ParseUint(buf.String(), 10, 64)
			if err != nil {
				toks[0].token = ILLEGAL
			} else {
				toks[0].token = NUMBER
				toks[0].value = num
			}
		}
	}
	return toks
}

func (l *lexer) consumeIdent() []*lexToken {
	buf := bytes.Buffer{}
	for {
		b, err := l.r.ReadByte()
		if err != nil {
			return nil
		}
		if isWhitespace(b) {
			l.r.UnreadByte()
			break
		}
		buf.WriteByte(b)
	}
	var tok *lexToken
	s := buf.String()
	if s == "true" {
		tok = &lexToken{token: BOOLEAN, value: true}
	} else if s == "false" {
		tok = &lexToken{token: BOOLEAN, value: false}
	} else {
		tok = &lexToken{token: lookup(buf.String()), value: buf.String()}
	}
	return []*lexToken{tok}
}

func isNumber(b byte) bool     { return unicode.IsDigit(rune(b)) }
func isLetter(b byte) bool     { return unicode.IsLetter(rune(b)) }
func isWhitespace(b byte) bool { return unicode.IsSpace(rune(b)) }
