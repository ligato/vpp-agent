package localdeps

import (
	"github.com/ligato/cn-infra/config"
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

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime. This is default empty implementation to not bother plugins
// that does not need to implement this method.
func (plugin *PluginLogDeps) Close() error {
	return nil
}

// PluginInfraDeps is standard set of plugin dependencies that
// will need probably every connector to DB/Messaging:
// - to report/write plugin status to StatusCheck
// - to know micro-service label prefix
type PluginInfraDeps struct {
	PluginLogDeps                               // inject
	config.PluginConfig                         // inject
	StatusCheck  statuscheck.PluginStatusWriter // inject
	ServiceLabel servicelabel.ReaderAPI         // inject
}
