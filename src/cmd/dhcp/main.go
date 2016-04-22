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
		os.Exit(1)
	}
	config.Print()

	// l := logger.New("dhcp")
	// l.NoFile()
	//
	// handler := dhcp.NewDHCPServer(config, nil)
	// handler.Serve()
}
