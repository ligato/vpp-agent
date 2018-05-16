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
		Program:        string(bytes.SplitN(reply.Program, []byte{0x00}, 2)[0]),
		Version:        string(bytes.SplitN(reply.Version, []byte{0x00}, 2)[0]),
		BuildDate:      string(bytes.SplitN(reply.BuildDate, []byte{0x00}, 2)[0]),
		BuildDirectory: string(bytes.SplitN(reply.BuildDirectory, []byte{0x00}, 2)[0]),
	}
	return info, nil
}
