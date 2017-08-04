package httpmux

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/rpc/grpcsync"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"github.com/unrolled/render"
	"net/http"
	"time"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "HTTP"

const (
	// DefaultHTTPPort is used during HTTP server startup unless different port was configured
	DefaultHTTPPort = "9191"
)

var (
	httpPort string
)

// init is here only for parsing program arguments
func init() {
	flag.StringVar(&httpPort, "http-port", DefaultHTTPPort,
		"Listen port for the Agent's HTTP server.")
}

// Plugin implements the Plugin interface.
type Plugin struct {
	LogFactory logging.LogFactory
	HTTPport   string

	logging.Logger
	server     *http.Server
	mx         *mux.Router
	formatter  *render.Render
	grpcServer *grpcsync.Adapter
}

// Init is entry point called by Agent Core
// - It prepares Gorilla MUX HTTP Router
// - registers grpc transport
func (plugin *Plugin) Init() (err error) {
	plugin.Logger, err = plugin.LogFactory.NewLogger(string(PluginID))
	if err != nil {
		return err
	}

	if plugin.HTTPport == "" {
		plugin.HTTPport = httpPort
	}

	plugin.mx = mux.NewRouter()
	plugin.formatter = render.New(render.Options{
		IndentJSON: true,
	})

	plugin.grpcServer = grpcsync.NewAdapter()
	plugin.Debug("grpctransp: ", plugin.grpcServer)
	err = datasync.RegisterTransport(&syncbase.Adapter{Watcher: plugin.grpcServer})

	return err
}

// RegisterHTTPHandler propagates to Gorilla mux
func (plugin *Plugin) RegisterHTTPHandler(path string,
	handler func(formatter *render.Render) http.HandlerFunc,
	methods ...string) *mux.Route {
	return plugin.mx.HandleFunc(path, handler(plugin.formatter)).Methods(methods...)
}

// AfterInit starts the HTTP server
func (plugin *Plugin) AfterInit() error {
	address := fmt.Sprintf("0.0.0.0:%s", plugin.HTTPport)
	//TODO NICE-to-HAVE make this configurable
	plugin.server = &http.Server{Addr: address, Handler: plugin.mx}

	var errCh chan error
	go func() {
		plugin.Info("Listening on http://", address)

		if err := plugin.server.ListenAndServe(); err != nil {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case err := <-errCh:
		return err
		// Wait 100ms to create a new stream, so it doesn't bring too much
		// overhead when retry.
	case <-time.After(100 * time.Millisecond):
		//everything is probably fine
		return nil
	}
}

// Close cleans up the resources
func (plugin *Plugin) Close() error {
	_, err := safeclose.CloseAll(plugin.grpcServer, plugin.server)
	return err
}
