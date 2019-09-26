package types

// ErrorResponse represents an error.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ComponentVersion describes the version information for a specific component.
type ComponentVersion struct {
	Name    string
	Version string
	Details map[string]string `json:",omitempty"`
}

// Version contains response of Engine API:
// GET "/version"
type Version struct {
	Components []ComponentVersion

	Version       string
	APIVersion    string
	MinAPIVersion string
	GitCommit     string
	GoVersion     string
	Os            string
	Arch          string
	KernelVersion string
	BuildTime     string
}

// Ping contains response of Engine API:
// GET "/_ping"
type Ping struct {
	APIVersion string
	OSType     string
}

type LoggerListOptions struct {
	Name string
}

type Logger struct {
	Logger string `json:"logger,omitempty"`
	Level  string `json:"level,omitempty"`
}
