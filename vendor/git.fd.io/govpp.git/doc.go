// Package govpp provides the entry point to govpp functionality. It provides the API for connecting the govpp core
// to VPP either using the default VPP adapter, or using the adapter previously set by SetAdapter function
// (useful mostly just for unit/integration tests with mocked VPP adapter).
//
// To create a connection to VPP, use govpp.Connect function:
//
//	conn, err := govpp.Connect()
//	if err != nil {
//		// handle error!
//	}
//	defer conn.Disconnect()
//
// Make sure you close the connection after using it. If the connection is not closed, it will leak resources. Please
// note that only one VPP connection is allowed for a single process.
//
// In case you need to mock the connection to VPP (e.g. for testing), use the govpp.SetAdapter function before
// calling govpp.Connect.
//
// Once connected to VPP, use the functions from the api package to communicate with it.
package govpp
