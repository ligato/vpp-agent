package aclplugin

import (
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/aclplugin/model/acl"
)

// Resync writes ACLs to the empty VPP
func (plugin *ACLConfigurator) Resync(acls []*acl.AccessLists_Acl) error {
	log.Debug("Resync ACLs started")

	var wasError error

	// Create VPP ACLs
	log.Debugf("Configuring %v new ACLs", len(acls))
	for _, aclInput := range acls {
		err := plugin.ConfigureACL(aclInput)
		if err != nil {
			wasError = err
		}
	}

	log.WithField("cfg", plugin).Debug("RESYNC ACLs end. ", wasError)

	return wasError
}
