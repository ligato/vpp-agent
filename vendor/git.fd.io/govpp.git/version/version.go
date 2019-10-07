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

package version

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var (
	name        = "govpp"
	version     = "v0.0.0+unknown"
	commitHash  = "unknown"
	buildBranch = "HEAD"
	buildStamp  = ""
	buildUser   = ""
	buildHost   = ""

	buildDate time.Time
)

func init() {
	buildstampInt64, _ := strconv.ParseInt(buildStamp, 10, 64)
	if buildstampInt64 == 0 {
		buildstampInt64 = time.Now().Unix()
	}
	buildDate = time.Unix(buildstampInt64, 0)
}

func Info() string {
	return fmt.Sprintf(`%s %s`, name, version)
}

func Verbose() string {
	return fmt.Sprintf(`%s
  Version:      %s
  Branch:   	%s
  Revision: 	%s
  Built by:  	%s@%s 
  Build date:	%s
  Go runtime:	%s (%s/%s)`,
		name,
		version, buildBranch, commitHash,
		buildUser, buildHost, buildDate.Format(time.UnixDate),
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}
