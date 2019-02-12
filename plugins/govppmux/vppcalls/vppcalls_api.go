package vppcalls

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging"
)

// VersionInfo contains values returned from ShowVersion
type VersionInfo struct {
	Program        string
	Version        string
	BuildDate      string
	BuildDirectory string
}

// VpeInfo contains information about VPP connection and process.
type VpeInfo struct {
	PID            uint32
	ClientIdx      uint32
	ModuleVersions []ModuleVersion
}

// ModuleVersion contains info about version of particular VPP module.
type ModuleVersion struct {
	Name  string
	Major uint32
	Minor uint32
	Patch uint32
}

func (m ModuleVersion) String() string {
	return fmt.Sprintf("%s-%d.%d.%d", m.Name, m.Major, m.Minor, m.Patch)
}

// VpeVppAPI provides methods for retrieving info and running CLI commands.
type VpeVppAPI interface {
	GetVersionInfo() (*VersionInfo, error)
	GetVpeInfo() (*VpeInfo, error)
	RunCli(cmd string) (string, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel) VpeVppAPI
}

func CompatibleVpeHandler(ch govppapi.Channel) VpeVppAPI {
	for ver, h := range Versions {
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			log.Debugf("version %s not compatible", ver)
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch)
	}
	panic("no compatible version available")
}
