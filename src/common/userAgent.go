// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"regexp"
	"strings"
)

var (
	userAgentRegex  = regexp.MustCompile(`^Mozilla\/\d\.0 \((.*?)\)`)
	macVersionRegex = regexp.MustCompile(`(\d{1,2}[_\.]\d{1,2}(?:[_\.]\d{1,2})?)`)
	// The trailing period is to make each key 5 characters long
	osXVersionNames = map[string]string{
		"10.5.": "Leopard",
		"10.6.": "Snow Leopard",
		"10.7.": "Lion",
		"10.8.": "Mountain Lion",
		"10.9.": "Mavericks",
		"10.10": "Yosemite",
		"10.11": "El Capitan",
		"10.12": "Sierra",
		"10.13": "High Sierra",
		"10.14": "Mojave",
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
	} else if strings.Contains(ua, "Kindle") {
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
