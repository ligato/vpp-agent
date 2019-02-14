package vpp1810

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/pkg/idxvpp"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
)

func init() {
	vppcalls.Versions["vpp1810"] = vppcalls.HandlerVersion{
		Msgs: l2ba.Messages,
		New: func(ch govppapi.Channel,
			ifIdx ifaceidx.IfaceMetadataIndex, bdIdx idxvpp.NameToIndex, log logging.Logger,
		) vppcalls.L2VppAPI {
			return NewL2VppHandler(ch, ifIdx, bdIdx, log)
		},
	}
}

type L2VppHandler struct {
	*BridgeDomainVppHandler
	*FIBVppHandler
	*XConnectVppHandler
}

func NewL2VppHandler(ch govppapi.Channel,
	ifIdx ifaceidx.IfaceMetadataIndex, bdIdx idxvpp.NameToIndex, log logging.Logger,
) *L2VppHandler {
	return &L2VppHandler{
		BridgeDomainVppHandler: newBridgeDomainVppHandler(ch, ifIdx, log),
		FIBVppHandler:          newFIBVppHandler(ch, ifIdx, bdIdx, log),
		XConnectVppHandler:     newXConnectVppHandler(ch, ifIdx, log),
	}
}

// BridgeDomainVppHandler is accessor for bridge domain-related vppcalls methods.
type BridgeDomainVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// FIBVppHandler is accessor for FIB-related vppcalls methods.
type FIBVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	bdIndexes    idxvpp.NameToIndex
	log          logging.Logger
}

// XConnectVppHandler is accessor for cross-connect-related vppcalls methods.
type XConnectVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewBridgeDomainVppHandler creates new instance of bridge domain vppcalls handler.
func newBridgeDomainVppHandler(callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger,
) *BridgeDomainVppHandler {
	return &BridgeDomainVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}

// NewFIBVppHandler creates new instance of FIB vppcalls handler.
func newFIBVppHandler(callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex, bdIndexes idxvpp.NameToIndex, log logging.Logger,
) *FIBVppHandler {
	return &FIBVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		bdIndexes:    bdIndexes,
		log:          log,
	}
}

// NewXConnectVppHandler creates new instance of cross connect vppcalls handler.
func newXConnectVppHandler(callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger,
) *XConnectVppHandler {
	return &XConnectVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}
