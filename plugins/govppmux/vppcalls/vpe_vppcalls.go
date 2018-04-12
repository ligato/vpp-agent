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
		Program:        string(reply.Program[:bytes.IndexByte(reply.Program, 0)]),
		Version:        string(reply.Version[:bytes.IndexByte(reply.Version, 0)]),
		BuildDate:      string(reply.BuildDate[:bytes.IndexByte(reply.BuildDate, 0)]),
		BuildDirectory: string(reply.BuildDirectory[:bytes.IndexByte(reply.BuildDirectory, 0)]),
	}
	return info, nil
}
