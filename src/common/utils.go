// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"errors"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	TimeFormat string = "2006-01-02 15:04"

	secondsInMinute int = 60
	secondsInHour   int = 60 * 60
)

var (
	userAgentRegex  = regexp.MustCompile(`^Mozilla\/\d\.0 \((.*?)\)`)
	macVersionRegex = regexp.MustCompile(`(\d{1,2}[_\.]\d{1,2}(?:[_\.]\d{1,2})?)`)
	// The trailing underscore is to make each key 5 characters long
	osXVersionNames = map[string]string{
		"10.5.": "Leopard",
		"10.6.": "Snow Leopard",
		"10.7.": "Lion",
		"10.8.": "Mountain Lion",
		"10.9.": "Mavericks",
		"10.10": "Yosemite",
		"10.11": "El Capitan",
	}

	windowsVersionNames = map[string]string{
		"5.1":  "XP",
		"5.2":  "XP",
		"6.0":  "Vista",
		"6.1":  "7",
		"6.2":  "8",
		"6.3":  "8.1",
		"10.0": "10",
	}
)

// StringInSlice searches a slice for a string
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ConvertToInt converts s to an int and ignores errors
func ConvertToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// FormatMacAddress will attempt to format and parse a string as a MAC address
func FormatMacAddress(mac string) (net.HardwareAddr, error) {
	// If no punctuation was provided, use the format xxxx.xxxx.xxxx
	if len(mac) == 12 {
		mac = mac[0:4] + "." + mac[4:8] + "." + mac[8:12]
	}
	return net.ParseMAC(mac)
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func ParseTime(time string) (int64, error) {
	clock := strings.Split(time, ":")
	if len(clock) != 2 {
		return 0, errors.New("Invalid time format. Expected HH:mm")
	}

	hours, err := strconv.Atoi(clock[0])
	if err != nil {
		return 0, errors.New("Hours is not a number")
	}
	minutes, err := strconv.Atoi(clock[1])
	if err != nil {
		return 0, errors.New("Minutes is not a number")
	}

	if hours < 0 || hours > 24 {
		return 0, errors.New("Hours must be between 0 and 24")
	}
	if minutes < 0 || minutes > 59 {
		return 0, errors.New("Minutes must be between 0 and 59")
	}

	return int64((hours * secondsInHour) + (minutes * secondsInMinute)), nil
}

func ParseUserAgent(ua string) string {
	parsedUA := ""
	match := userAgentRegex.FindStringSubmatch(ua)
	if len(match) != 2 {
		return ""
	}

	parts := strings.Split(match[1], "; ")

	if strings.HasPrefix(parts[0], "Windows") {
		parsedUA = parseWindowsUA(parts)
	} else if strings.HasPrefix(parts[0], "Mac") {
		parsedUA = parseMacUA(parts)
	} else if strings.HasPrefix(parts[0], "iP") { // iPad, iPhone, iPod
		parsedUA = parseiOSUA(parts)
	} else if strings.Index(ua, "Kindle") != -1 {
		parsedUA = "Kindle"
	} else if parts[0] == "X11" {
		parsedUA = parseLinuxUA(parts)
	} else if parts[0] == "Linux" || strings.HasPrefix(parts[0], "Android") {
		parsedUA = parseAndroidUA(parts)
	} else if strings.HasPrefix(parts[0], "Fedora") {
		parsedUA = "Fedora"
	}

	return parsedUA
}

func parseWindowsUA(ua []string) string {
	uStr := "Windows"
	verStr := ""
	for _, s := range ua {
		if strings.HasPrefix(s, "Windows NT") {
			verStr = s
			break
		}
	}
	if verStr != "" {
		if ver, ok := windowsVersionNames[verStr[11:]]; ok {
			uStr += " " + ver
		}
	}

	return uStr
}

func parseAndroidUA(ua []string) string {
	uStr := "Android"
	for _, s := range ua {
		if strings.HasPrefix(s, "Android") {
			uStr = s
			break
		}
	}
	return uStr
}

func parseiOSUA(ua []string) string {
	uStr := ua[0]
	verStr := ""
	for _, s := range ua {
		if strings.HasPrefix(s, "CPU") {
			verStr = s
			break
		}
	}
	if verStr != "" {
		match := macVersionRegex.FindStringSubmatch(verStr)
		if len(match) > 1 {
			uStr += " iOS " + strings.Replace(match[1], "_", ".", -1)
		}
	}
	return uStr
}

func parseMacUA(ua []string) string {
	uStr := "Mac"
	if len(ua) < 2 {
		return uStr
	}
	verStr := ""
	for _, uaPart := range ua[1:] { // The first is always "Macintosh"
		if strings.HasPrefix(uaPart, "Intel") && len(uaPart) >= 15 {
			verStr = uaPart[15:]
			break
		} else if strings.HasPrefix(uaPart, "PPC") && len(uaPart) >= 13 {
			verStr = uaPart[13:]
			break
		}
	}
	if verStr != "" {
		verStr = strings.Replace(verStr, "_", ".", -1)
		uStr = "Mac OS X " + verStr
		if len(verStr) >= 4 {
			if len(verStr) == 4 {
				verStr += "."
			}
			if name, ok := osXVersionNames[verStr[:5]]; ok {
				uStr += " (" + name + ")"
			}
		}
	}
	return uStr
}

func parseLinuxUA(ua []string) string {
	if len(ua) < 2 {
		return "X11"
	}

	// We only look at the first piece. If it's specific we say the distro
	// Otherwise just a generic "it's Linux"
	if strings.HasPrefix(ua[1], "CrOS") {
		return "Chromium OS"
	} else if strings.HasPrefix(ua[1], "Linux i") {
		return "Linux 32 bit"
	} else if strings.HasPrefix(ua[1], "Linux x") {
		return "Linux 64 bit"
	} else if strings.HasPrefix(ua[1], "Ubuntu") {
		return "Ubuntu"
	} else if strings.HasPrefix(ua[1], "Fedora") {
		return "Fedora"
	}

	return "Linux"
}
