package vpp1908

import (
	"errors"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
)

// AddGtpuTunnel adds new GTPU interface.
func (h *InterfaceVppHandler) AddGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink, multicastIf uint32) (uint32, error) {
    err := errors.New("Not implemented")
    return 0, err
}

// DelGtpuTunnel removes GTPU interface.
func (h *InterfaceVppHandler) DelGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink) error {
    err := errors.New("Not implemented")
	return err
}
