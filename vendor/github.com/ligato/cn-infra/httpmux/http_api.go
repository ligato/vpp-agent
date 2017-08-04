package httpmux

import (
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"net/http"
)

// HTTPHandlers is an interface that is useful for other plugins that need to register HTTP Handlers.
// Use this interface as type for the field in terms of dependency injection.
type HTTPHandlers interface {
	// RegisterHTTPHandler propagates to Gorilla mux
	RegisterHTTPHandler(path string,
		handler func(formatter *render.Render) http.HandlerFunc,
		methods ...string) *mux.Route
}
