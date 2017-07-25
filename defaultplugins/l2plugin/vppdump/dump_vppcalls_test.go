package vppdump

import (
	"fmt"
	"os"
	"testing"

	"git.fd.io/govpp.git"
)

func TestDumpL2(t *testing.T) {
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

	res, err := DumpBridgeDomains(ch)
	fmt.Printf("%+v\n", res)

	res2, err := DumpFIBTableEntries(ch)
	fmt.Printf("%+v\n", res2)
	for _, fib := range res2 {
		fmt.Printf("%+v\n", fib)
	}

	res3, _ := DumpXConnectPairs(ch)
	fmt.Printf("%+v\n", res3)
	for _, xconn := range res3 {
		fmt.Printf("%+v\n", xconn)
	}
}
