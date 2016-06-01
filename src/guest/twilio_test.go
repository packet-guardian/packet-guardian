package guest

import "testing"

type phoneTest struct {
	input    string
	expected string
}

var phoneTests = []phoneTest{
	phoneTest{"555-124-5785", "5551245785"},
	phoneTest{"5551245785", "5551245785"},
	phoneTest{"555-124-57853", ""},
	phoneTest{"", ""},
	phoneTest{"abc", ""},
	phoneTest{"abcdefghij", ""},
	phoneTest{"1-555-124-5785", "5551245785"},
	phoneTest{"2-555-124-5785", ""},
	phoneTest{"15551245785", "5551245785"},
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
