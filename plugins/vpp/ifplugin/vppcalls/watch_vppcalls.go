package vppcalls

import (
	"os"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/go-errors/errors"
	if_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
)

// WatchInterfaceEvents starts watching for interface events.
func (h *IfVppHandler) WatchInterfaceEvents(events chan<- *InterfaceEvent) error {
	notifChan := make(chan govppapi.Message)

	// subscribe for receiving SwInterfaceEvents notifications
	_, err := h.callsChannel.SubscribeNotification(notifChan, &if_api.SwInterfaceEvent{})
	if err != nil {
		return errors.Errorf("failed to subscribe VPP notification (sw_interface_event): %v", err)
	}

	go func() {
		for {
			select {
			case e := <-notifChan:
				ifEvent, ok := e.(*if_api.SwInterfaceEvent)
				if !ok {
					continue
				}
				events <- &InterfaceEvent{
					SwIfIndex:  ifEvent.SwIfIndex,
					AdminState: ifEvent.AdminUpDown,
					LinkState:  ifEvent.LinkUpDown,
					Deleted:    ifEvent.Deleted != 0,
				}
			}
		}
	}()

	// enable interface state notifications from VPP
	wantIfEventsReply := &if_api.WantInterfaceEventsReply{}
	err = h.callsChannel.SendRequest(&if_api.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantIfEventsReply)
	if err != nil {
		return errors.Errorf("failed to watch interface events: %v", err)
	}

	return nil
}
