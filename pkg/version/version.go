//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package version provides information about app version.
package version

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var (
	app       = "vpp-agent"
	version   = "v3.1.0"
	gitCommit = "unknown"
	gitBranch = "HEAD"
	buildUser = "unknown"
	buildHost = "unknown"
	buildDate = ""
)

var buildTime time.Time
var revision string

func init() {
	if buildDate != "" {
		buildstampInt64, _ := strconv.ParseInt(buildDate, 10, 64)
		buildTime = time.Unix(buildstampInt64, 0)
	}
	revision = gitCommit
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if gitBranch != "HEAD" {
		revision += fmt.Sprintf("@%s", gitBranch)
	}
}

// App returns app name.
func App() string {
	return app
}

// Version returns version string.
func Version() string {
	return version
}

// Data returns version data.
func Data() (ver, rev, date string) {
	return version, revision, buildTime.Format(time.UnixDate)
}

func Short() string {
	return fmt.Sprintf(`%s %s`, app, version)
}

func BuiltOn() string {
	stamp := buildTime.Format(time.UnixDate)
	if !buildTime.IsZero() {
		stamp += fmt.Sprintf(" (%s)", timeAgo(buildTime))
	}
	return stamp
}

func BuiltBy() string {
	return fmt.Sprintf("%s@%s (%s %s/%s)",
		buildUser, buildHost, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

// Info returns string with complete version info on single line.
func Info() string {
	return fmt.Sprintf(`%s %s (%s) built by %s@%s on %v`,
		app, version, revision, buildUser, buildHost, BuiltOn(),
	)
}

// Detail returns string with detailed version info on separate lines.
func Detail() string {
	return fmt.Sprintf(`%s
  Version:   	%s
  Branch:   	%s
  Revision:  	%s
  Built By:  	%s@%s 
  Build Date:	%s
  Go Runtime:	%s (%s/%s)`,
		app, version, gitBranch, revision,
		buildUser, buildHost, buildTime.Format(time.UnixDate),
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

func timeAgo(t time.Time) string {
	const timeDay = time.Hour * 24
	if ago := time.Since(t); ago > timeDay {
		return fmt.Sprintf("%v days ago", float64(ago.Round(timeDay)/timeDay))
	} else if ago > time.Hour {
		return fmt.Sprintf("%v hours ago", ago.Round(time.Hour).Hours())
	} else if ago > time.Minute {
		return fmt.Sprintf("%v minutes ago", ago.Round(time.Minute).Minutes())
	}
	return "just now"
}
