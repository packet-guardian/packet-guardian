// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"strconv"
)

type token int

type lexToken struct {
	token     token
	value     interface{}
	line, pos int
}

const (
	ANY token = -1

	ILLEGAL token = iota
	EOF
	COMMENT
	EOL

	literal_beg
	NUMBER
	STRING
	IP_ADDRESS
	BOOLEAN
	literal_end

	keyword_beg
	END
	GLOBAL
	NETWORK
	SUBNET
	POOL
	REGISTERED
	UNREGISTERED
	SERVER_IDENTIFIER
	RANGE
	INCLUDE
	LOCAL
	IGNORE_REGISTRATION
	DECL_OPTION
	CODE
	OPTION_TYPE

	setting_beg
	OPTION
	FREE_LEASE_AFTER
	DEFAULT_LEASE_TIME
	MAX_LEASE_TIME
	setting_end
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	NUMBER:     "NUMBER",
	STRING:     "STRING",
	IP_ADDRESS: "IP_ADDRESS",
	BOOLEAN:    "BOOLEAN",

	END:                 "end",
	GLOBAL:              "global",
	NETWORK:             "network",
	SUBNET:              "subnet",
	POOL:                "pool",
	REGISTERED:          "registered",
	UNREGISTERED:        "unregistered",
	SERVER_IDENTIFIER:   "server-identifier",
	RANGE:               "range",
	INCLUDE:             "include",
	LOCAL:               "local",
	IGNORE_REGISTRATION: "ignore-registration",
	DECL_OPTION:         "decloption",
	CODE:                "code",
	OPTION_TYPE:         "type",

	OPTION:             "option",
	FREE_LEASE_AFTER:   "free-lease-after",
	DEFAULT_LEASE_TIME: "default-lease-time",
	MAX_LEASE_TIME:     "max-lease-time",
}

var keywords map[string]token

func (tok token) string() string {
	s := ""
	if 0 <= tok && tok < token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

func (tok *lexToken) string() string {
	return fmt.Sprintf("%s: %v", tok.token.string(), tok.value)
}

func init() {
	keywords = make(map[string]token)
	for i := keyword_beg + 1; i < keyword_end-1; i++ {
		keywords[tokens[i]] = i
	}
}

func lookup(ident string) token {
	if tok, valid := keywords[ident]; valid {
		return tok
	}
	return STRING
}

func (tok token) isSetting() bool { return setting_beg < tok && tok < setting_end }
