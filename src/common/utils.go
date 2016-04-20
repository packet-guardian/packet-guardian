package common

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	TimeFormat string = "2006-01-02 15:04"

	secondsInMinute int = 60
	secondsInHour   int = 60 * 60
)

// NotImplementedHandler is a mock handler for paths that aren't implemented yet
func NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf("The path \"%s\" is not implemented yet\n", r.URL.Path)
	w.Write([]byte(resp))
}

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
