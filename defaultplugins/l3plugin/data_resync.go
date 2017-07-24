package l3plugin

import (
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/l3plugin/model/l3"
)

// Resync confgures the empty VPP (overwrites the static route)
func (plugin *RouteConfigurator) Resync(staticRoutes *l3.StaticRoutes) error {
	log.WithField("cfg", plugin).Debug("RESYNC routes begin. ")
	// TODO lookup vpp Route Configs

	if staticRoutes != nil {

		wasError := plugin.ConfigureRoutes(staticRoutes)
		log.WithField("cfg", plugin).Debug("RESYNC routes end. ", wasError)
		return wasError
	}

	return nil
}
