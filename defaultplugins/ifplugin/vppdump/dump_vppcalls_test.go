package vppdump

import (
	"fmt"
	"os"
	"testing"

	"git.fd.io/govpp.git"
)

func TestDumpInterfaces(t *testing.T) {
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

	res, _ := DumpInterfaces(ch)
	fmt.Println(res)
}
