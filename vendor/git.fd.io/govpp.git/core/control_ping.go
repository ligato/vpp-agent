package core

import "git.fd.io/govpp.git/api"

var (
	msgControlPing      api.Message = new(ControlPing)
	msgControlPingReply api.Message = new(ControlPingReply)
)

// SetControlPing sets the control ping message used by core.
func SetControlPing(m api.Message) {
	msgControlPing = m
}

// SetControlPingReply sets the control ping reply message used by core.
func SetControlPingReply(m api.Message) {
	msgControlPingReply = m
}

type ControlPing struct{}

func (*ControlPing) GetMessageName() string {
	return "control_ping"
}
func (*ControlPing) GetCrcString() string {
	return "51077d14"
}
func (*ControlPing) GetMessageType() api.MessageType {
	return api.RequestMessage
}

type ControlPingReply struct {
	Retval      int32
	ClientIndex uint32
	VpePID      uint32
}

func (*ControlPingReply) GetMessageName() string {
	return "control_ping_reply"
}
func (*ControlPingReply) GetCrcString() string {
	return "f6b0b8ca"
}
func (*ControlPingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func init() {
	api.RegisterMessage((*ControlPing)(nil), "ControlPing")
	api.RegisterMessage((*ControlPingReply)(nil), "ControlPingReply")
}
