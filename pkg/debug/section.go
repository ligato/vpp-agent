package debug

import (
	"os"
	"strings"
)

const envDebug = "DEBUG"

func init() {
	if os.Getenv("DEBUG_ENABLED") != "" {
		Enable()
	}
}

// IsEnabled checks whether the debug flag is set or not.
func IsEnabled() bool {
	return os.Getenv(envDebug) != ""
}

// Enable sets the DEBUG env var to true.
func Enable() {
	if IsEnabled() {
		return
	}
	os.Setenv(envDebug, "1")
}

// Disable sets the DEBUG env var to false.
func Disable() {
	os.Setenv(envDebug, "")
}

// IsEnabledFor returns true if DEBUG env var contains all sections.
func IsEnabledFor(sections ...string) bool {
	env := os.Getenv(envDebug)
	if env == "" {
		return false
	}
	for _, s := range sections {
		if !strings.Contains(env, s) {
			return false
		}
	}
	return true
}
