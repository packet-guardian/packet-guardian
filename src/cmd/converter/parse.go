package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/onesimus-systems/packet-guardian/src/common"
	"github.com/onesimus-systems/packet-guardian/src/models"
)

var devices = make(map[string]*models.Device)

func parseFile(path string, e *common.Environment) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return parse(bufio.NewScanner(file), e)
}

func parse(scan *bufio.Scanner, e *common.Environment) error {
	var user *models.User
	controlUser := false
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		if line[0] == '%' { // Control line
			if string(line) == "% stop" {
				return nil
			}

			username := ""
			manage := true
			limit := models.UserDeviceLimitGlobal
			var devExp *models.UserDeviceExpiration
			parts := strings.Split(line, " ")[1:]
			for _, part := range parts {
				setVal := strings.SplitN(part, "=", 2)
				setting := setVal[0]
				val := setVal[1]
				switch setting {
				case "user":
					username = val
					break
				case "expiration":
					if val == "global" {
						devExp = &models.UserDeviceExpiration{Mode: models.UserDeviceExpirationGlobal}
					} else if val == "rolling" {
						devExp = &models.UserDeviceExpiration{Mode: models.UserDeviceExpirationRolling}
					} else {
						devExp = &models.UserDeviceExpiration{Mode: models.UserDeviceExpirationNever}
					}
					break
				case "manage":
					manage = (val == "true")
					break
				case "limit":
					if val == "unlimited" {
						limit = models.UserDeviceLimitUnlimited
					} else {
						limit = models.UserDeviceLimitGlobal
					}
				}
			}

			user, _ = models.GetUserByUsername(e, username)
			user.DeviceExpiration = devExp
			user.CanManage = manage
			user.DeviceLimit = limit
			controlUser = (username != "")
			continue
		}

		if user == nil {
			return errors.New("-- At least one control line must exist")
		}

		var err error
		parts := strings.Split(line, "#")
		registeredFrom := net.ParseIP(parts[3])
		dateRegistered, err := time.ParseInLocation("20060102 15:04:05", parts[2], time.Local)
		if err != nil {
			fmt.Printf("-- Error parsing date: %v\n", err)
			continue
		}
		userAgent := parts[1]

		hostPart := strings.SplitN(parts[0], " ", 7)
		description := hostPart[1]
		mac, _ := net.ParseMAC(hostPart[5])
		platform := common.ParseUserAgent(userAgent)
		if platform == "" {
			if strings.Index(userAgent, "(Manual") != -1 {
				platform = strings.SplitN(userAgent, " ", 2)[0]
			}
		}

		if user.Username == "" {
			user.Username = hostPart[1][:strings.LastIndex(hostPart[1], "-")]
		}

		if dev, ok := devices[mac.String()]; ok {
			if dev.Username != user.Username {
				if dev.Username == "library" {
					fmt.Printf("-- Warning: User override for MAC %s: library from %s\n", mac.String(), user.Username)
				} else if dev.Username == "housing" || user.Username == "housing" {
					dev.Username = "housing"
					fmt.Printf("-- Warning: User override for MAC %s: housing from %s\n", mac.String(), user.Username)
				} else if dev.Username == "mewirelss" {
					dev.Username = user.Username
					fmt.Printf("-- Warning: User override 'mewireless' for MAC %s: %s\n", mac.String(), user.Username)
				} else if dev.Username == "lawilson" {
					fmt.Printf("-- Warning: User override for MAC %s: lawilson from %s\n", mac.String(), user.Username)
				} else {
					fmt.Printf(
						"-- Warning: Conflicting usernames for MAC %s: %s and %s not included\n",
						mac.String(),
						dev.Username,
						user.Username,
					)
					delete(devices, mac.String())
				}
			}
		} else {
			dev := models.NewDevice(nil)
			dev.MAC = mac
			dev.Username = user.Username
			dev.RegisteredFrom = registeredFrom
			dev.Platform = platform
			dev.Expires = user.DeviceExpiration.NextExpiration(e)
			dev.DateRegistered = dateRegistered
			dev.UserAgent = userAgent
			dev.Description = description
			devices[mac.String()] = dev
		}

		if controlUser {
			user.Save()
		} else {
			user.Username = ""
		}
	}
	return scan.Err()
}

func writeOutput() {
	fmt.Println("PRAGMA synchronous = OFF;")
	fmt.Println("PRAGMA journal_mode = MEMORY;")
	fmt.Println("BEGIN TRANSACTION;")
	for _, dev := range devices {
		fmt.Printf(
			`INSERT INTO "device" ("mac", "username", "registered_from", "platform", "expires", "date_registered", "user_agent", "description") VALUES ('%s','%s','%s','%s','%d','%d','%s','%s');%s`,
			dev.MAC.String(),
			dev.Username,
			dev.RegisteredFrom.String(),
			dev.Platform,
			dev.Expires.Unix(),
			dev.DateRegistered.Unix(),
			dev.UserAgent,
			dev.Description,
			"\n",
		)
	}
	fmt.Println("END TRANSACTION;")
}
