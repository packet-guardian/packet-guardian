package main

import (
	"fmt"
	"os"

	"github.com/onesimus-systems/packet-guardian/src/dhcp"
)

func main() {
	config, err := dhcp.ParseFile(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
	}
	config.Print()
}
