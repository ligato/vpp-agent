package vppcalls

import (
	"bytes"
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
)

// VersionInfo contains values returned from ShowVersion
type VersionInfo struct {
	Program        string
	Version        string
	BuildDate      string
	BuildDirectory string
}

// GetVersionInfo retrieves version information
func GetVersionInfo(vppChan *govppapi.Channel) (*VersionInfo, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}
	if reply.Retval != 0 {
		return nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	info := &VersionInfo{
		Program:        string(bytes.Trim(reply.Program, "\x00")),
		Version:        string(bytes.Trim(reply.Version, "\x00")),
		BuildDate:      string(bytes.Trim(reply.BuildDate, "\x00")),
		BuildDirectory: string(bytes.Trim(reply.BuildDirectory, "\x00")),
	}
	return info, nil
}
