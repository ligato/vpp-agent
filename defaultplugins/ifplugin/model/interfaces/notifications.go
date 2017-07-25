package interfaces

// InterfaceStateNotificationType is type of the notification
type InterfaceStateNotificationType int32

const (
	// UNKNOWN is default type
	UNKNOWN InterfaceStateNotificationType = 0
	// UPDOWN represents Link UP/DOWN notification
	UPDOWN InterfaceStateNotificationType = 1
	// COUNTERS represents interface state with updated counters
	COUNTERS InterfaceStateNotificationType = 2
	// DELETED represents the event when the interface was deleted from the VPP
	// Note, some north bound config updates require delete and create the network interface one more time
	DELETED InterfaceStateNotificationType = 3
)

// InterfaceStateNotification aggregates status UP/DOWN/DELETED/UNKNOWN with the details (state) about the interfaces
// including counters
type InterfaceStateNotification struct {
	// Type of the notification
	Type InterfaceStateNotificationType
	// State of the network interface
	State *InterfacesState_Interface
}
