// Package core provides connectivity to VPP via the adapter: sends and receives the messages to/from VPP,
// marshalls/unmarshalls them and forwards them between the client Go channels and the VPP.
//
// The interface_plugin APIs the core exposes is tied to a connection: Connect provides a connection, that cane be
// later used to request an API channel via NewAPIChannel / NewAPIChannelBuffered functions:
//
//	conn, err := govpp.Connect()
//	if err != nil {
//		// handle error!
//	}
//	defer conn.Disconnect()
//
//	ch, err := conn.NewAPIChannel()
//	if err != nil {
//		// handle error!
//	}
//	defer ch.Close()
//
// Note that one application can open only one connection, that can serve multiple API channels.
package core
