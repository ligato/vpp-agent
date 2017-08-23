package localdeps

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
)

// PluginLogDeps is minimal set of plugin dependencies that
// will probably use every plugin to:
// - log using plugin logger or child (prefixed) logger (if plugin needs more than one)
// - to know the PluginName
type PluginLogDeps struct {
	Log        logging.PluginLogger //inject
	PluginName core.PluginName      //inject
}

// PluginInfraDeps is standard set of plugin dependencies that
// will need probably every connector to DB/Messaging:
// - to report/write plugin status to StatusCheck
// - to know micro-service label prefix
type PluginInfraDeps struct {
	PluginLogDeps                                //inject
	StatusCheck   statuscheck.PluginStatusWriter //inject
	ServiceLabel  servicelabel.ReaderAPI         //inject
}
