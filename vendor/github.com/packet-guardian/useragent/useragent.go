package useragent

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

// Names of various operating systems
const (
	MacOS     = "macOS"
	Windows   = "Windows"
	Linux     = "Linux"
	Ubuntu    = "Ubuntu"
	Fedora    = "Fedora"
	IPhone    = "iPhone"
	IPad      = "iPad"
	IPodTouch = "iPod touch"
	IOS       = "iOS"
	Kindle    = "Kindle"
	ChromeOS  = "ChromeOS"
	Android   = "Android"
	Unknown   = "Unknown"
)

// Windows versions
const (
	WindowsXP    = "XP"
	WindowsVista = "Vista"
	Windows7     = "7"
	Windows8     = "8"
	Windows81    = "8.1"
	Windows10    = "10"
)

// macOS version names
const (
	MacOSLeopard      = "Leopard"
	MacOSSnowLeopard  = "Snow Leopard"
	MacOSLion         = "Lion"
	MacOSMountainLion = "Mountain Lion"
	MacOSMavericks    = "Mavericks"
	MacOSYosemite     = "Yosemite"
	MacOSElCapitan    = "El Capitan"
	MacOSSierra       = "Sierra"
	MacOSHighSierra   = "High Sierra"
	MacOSMojave       = "Mojave"
)

// UserAgent represents the OS name and version information from a parsed user agent.
// Not every OS has a Distro, Version, or VersionName.
type UserAgent struct {
	OS          string // OS is the name of the operating system
	Distro      string // Distro is used for Linux and to differentiate iOS products
	Version     string // Version is a numerical version number
	VersionName string // VersionName or edition is the common name used for a specific OS version
}

func (ua UserAgent) String() string {
	s := strings.Builder{}
	s.WriteString(ua.OS)

	if ua.VersionName != "" {
		s.WriteByte(' ')
		s.WriteString(ua.VersionName)
	}

	if ua.Version != "" {
		s.WriteString(" (")
		s.WriteString(ua.Version)
		s.WriteByte(')')
	}

	if ua.Distro != "" {
		s.WriteString(" (")
		s.WriteString(ua.Distro)
		s.WriteByte(')')
	}

	return s.String()
}

// ParseUserAgent takes a user agent and returns OS name and version information.
func ParseUserAgent(ua string) UserAgent {
	match := userAgentRegex.FindStringSubmatch(ua)
	if len(match) != 2 {
		return UserAgent{OS: Unknown}
	}

	parts := strings.Split(match[1], "; ")

	if strings.HasPrefix(parts[0], "Windows") {
		return parseWindowsUA(parts)
	} else if strings.HasPrefix(parts[0], "Mac") {
		return parseMacUA(parts)
	} else if strings.HasPrefix(parts[0], "iP") { // iPad, iPhone, iPod
		return parseiOSUA(parts)
	} else if strings.Contains(ua, "Kindle") {
		return UserAgent{OS: Kindle}
	} else if parts[0] == "X11" {
		return parseLinuxUA(parts)
	} else if parts[0] == "Linux" || strings.HasPrefix(parts[0], "Android") {
		return parseAndroidUA(parts)
	} else if strings.HasPrefix(parts[0], "Fedora") {
		return UserAgent{OS: Linux, Distro: Fedora}
	}

	return UserAgent{OS: Unknown}
}

func parseWindowsUA(ua []string) UserAgent {
	verStr := ""
	for _, s := range ua {
		if strings.HasPrefix(s, "Windows NT") {
			verStr = s
			break
		}
	}
	if verStr != "" {
		version := verStr[11:]
		if edition, ok := windowsVersionNames[version]; ok {
			return UserAgent{OS: Windows, Version: version, VersionName: edition}
		}
	}

	return UserAgent{OS: Windows}
}

func parseAndroidUA(ua []string) UserAgent {
	for _, s := range ua {
		if strings.HasPrefix(s, "Android") {
			return UserAgent{OS: Android, Version: s[8:]}
		}
	}
	return UserAgent{OS: Android}
}

func parseiOSUA(ua []string) UserAgent {
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
			return UserAgent{OS: IOS, Distro: ua[0], Version: strings.Replace(match[1], "_", ".", -1)}
		}
	}
	return UserAgent{OS: IOS, Distro: ua[0]}
}

func parseMacUA(ua []string) UserAgent {
	if len(ua) < 2 {
		return UserAgent{OS: MacOS}
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
		version := verStr
		edition := ""
		if len(verStr) >= 4 {
			if len(verStr) == 4 {
				verStr += "."
			}
			if name, ok := osXVersionNames[verStr[:5]]; ok {
				edition = name
			}
		}
		return UserAgent{OS: MacOS, Version: version, VersionName: edition}
	}
	return UserAgent{OS: MacOS}
}

func parseLinuxUA(ua []string) UserAgent {
	if len(ua) >= 2 {
		// We only look at the first piece. If it's specific we say the distro
		// Otherwise just a generic "it's Linux"
		if strings.HasPrefix(ua[1], "CrOS") {
			return UserAgent{OS: ChromeOS}
		} else if strings.HasPrefix(ua[1], "Ubuntu") {
			return UserAgent{OS: Linux, Distro: Ubuntu}
		} else if strings.HasPrefix(ua[1], "Fedora") {
			return UserAgent{OS: Linux, Distro: Fedora}
		}
	}

	return UserAgent{OS: Linux}
}
