package vppcalls

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/ligato/cn-infra/logging"
)

// VersionInfo contains values returned from ShowVersion
type VersionInfo struct {
	Program        string
	Version        string
	BuildDate      string
	BuildDirectory string
}

// GetVersionInfo retrieves version information
func GetVersionInfo(log logging.Logger, vppChan *govppapi.Channel) (*VersionInfo, error) {
	log.Debugf("retrieving version info")

	req := new(vpe.ShowVersion)
	reply := new(vpe.ShowVersionReply)

	// Send message
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}
	if reply.Retval != 0 {
		return nil, fmt.Errorf("ShowVersionReply returned %d", reply.Retval)
	}

	info := &VersionInfo{
		Program:        string(reply.Program),
		Version:        string(reply.Version),
		BuildDate:      string(reply.BuildDate),
		BuildDirectory: string(reply.BuildDirectory),
	}
	return info, nil
}
