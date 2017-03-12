// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package guest

import "testing"

type phoneTest struct {
	input    string
	expected string
}

var phoneTests = []phoneTest{
	{"555-124-5785", "5551245785"},
	{"5551245785", "5551245785"},
	{"555-124-57853", ""},
	{"", ""},
	{"abc", ""},
	{"abcdefghij", ""},
	{"1-555-124-5785", "5551245785"},
	{"2-555-124-5785", ""},
	{"15551245785", "5551245785"},
}

func TestPhoneFormat(t *testing.T) {
	for _, test := range phoneTests {
		formatted, err := formatPhoneNumber(test.input)
		if test.expected == "" && err == nil {
			t.Fatalf("Expected error but didn't get one. %s", test.input)
		}

		if formatted != test.expected {
			t.Errorf("Incorrectly formatted phone number %s. Expected %s, got %s", test.input, test.expected, formatted)
		}
	}
}
