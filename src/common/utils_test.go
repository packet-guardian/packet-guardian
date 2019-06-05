package common

import (
	"bytes"
	"testing"
)

var stringInSliceTests = []struct {
	list     []string
	input    string
	expected bool
}{
	{
		list:     []string{"aa", "bb", "cc"},
		input:    "aa",
		expected: true,
	},
	{
		list:     []string{"aa", "bb", "cc"},
		input:    "a",
		expected: false,
	},
	{
		list:     []string{},
		input:    "aa",
		expected: false,
	},
	{
		list:     nil,
		input:    "aa",
		expected: false,
	},
	{
		list:     []string{"aa", "bb", "cc"},
		input:    "dd",
		expected: false,
	},
}

func TestStringInSlice(t *testing.T) {
	for _, testcase := range stringInSliceTests {
		if StringInSlice(testcase.input, testcase.list) != testcase.expected {
			t.Errorf("StringInSlice failed, expected %t, got %t", testcase.expected, !testcase.expected)
		}
	}
}

var convertToInt = []struct {
	input    string
	expected int
}{
	{
		input:    "-1",
		expected: -1,
	},
	{
		input:    "",
		expected: 0,
	},
	{
		input:    "0",
		expected: 0,
	},
	{
		input:    "aa",
		expected: 0,
	},
	{
		input:    "125346",
		expected: 125346,
	},
}

func TestConvertToInt(t *testing.T) {
	for _, testcase := range convertToInt {
		i := ConvertToInt(testcase.input)
		if i != testcase.expected {
			t.Errorf("ConvertToInt failed, expected %d, got %d", testcase.expected, i)
		}
	}
}

var formatMACAddressTests = []struct {
	input     string
	expected  []byte
	shouldErr bool
}{
	{
		input:    "1234567890ab",
		expected: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB},
	},
	{
		input:    "12:34:56:78:90:ab",
		expected: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB},
	},
	{
		input:    "1234.5678.90ab",
		expected: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB},
	},
	{
		input:     "1234567890abcdef",
		shouldErr: true,
	},
	{
		input:     "",
		shouldErr: true,
	},
}

func TestFormatMACAddress(t *testing.T) {
	for _, testcase := range formatMACAddressTests {
		i, err := FormatMacAddress(testcase.input)
		if err != nil {
			if !testcase.shouldErr {
				t.Errorf("FormatMacAddress errored, %q", err.Error())
			}
			continue
		}

		if testcase.shouldErr {
			t.Errorf("FormatMacAddress didn't return an error for input %s", testcase.input)
			continue
		}

		if !bytes.Equal(i, testcase.expected) {
			t.Errorf("FormatMacAddress failed, expected %q, got %q", testcase.expected, i)
		}
	}
}

var parseTimeIntervalTests = []struct {
	input     string
	expected  int64
	shouldErr bool
}{
	{
		input:     "",
		shouldErr: true,
	},
	{
		input:     ":",
		shouldErr: true,
	},
	{
		input:    "00:00",
		expected: 0,
	},
	{
		input:    "1:00",
		expected: 3600,
	},
	{
		input:     "25:00",
		shouldErr: true,
	},
	{
		input:    "24:00",
		expected: 24 * 60 * 60,
	},
	{
		input:    "24:30",
		expected: 24*60*60 + 30*60,
	},
}

func TestParseTime(t *testing.T) {
	for _, testcase := range parseTimeIntervalTests {
		i, err := ParseTime(testcase.input)
		if err != nil {
			if !testcase.shouldErr {
				t.Errorf("ParseTime errored, %q", err.Error())
			}
			continue
		}

		if testcase.shouldErr {
			t.Errorf("ParseTime didn't return an error for input %s", testcase.input)
			continue
		}

		if i != testcase.expected {
			t.Errorf("ParseTime failed, expected %d, got %d", testcase.expected, i)
		}
	}
}
