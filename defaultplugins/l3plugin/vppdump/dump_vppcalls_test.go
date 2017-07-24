package vppdump

import (
	"fmt"
	"os"
	"testing"

	"git.fd.io/govpp.git"
)

func TestDumpL3(t *testing.T) {
	// connect to VPP
	conn, err := govpp.Connect()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer ch.Close()

	res3, err := DumpStaticRoutes(ch)
	fmt.Printf("%+v\n", res3)
	for _, routes := range res3 {
		for _, route := range routes.IP {
			fmt.Printf("%+v %+v\n", route, route.NextHops)
		}
	}
}
