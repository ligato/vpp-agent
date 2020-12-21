// Copyright (c) 2020 Pantheon.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package contextdecorator handles insertions and extractions of orchestrator related data from context.
package contextdecorator

import "context"

type dataSrcKeyT string

var dataSrcKey = dataSrcKeyT("dataSrc")

func DataSrcContext(ctx context.Context, dataSrc string) context.Context {
	return context.WithValue(ctx, dataSrcKey, dataSrc)
}

func DataSrcFromContext(ctx context.Context) (dataSrc string, ok bool) {
	dataSrc, ok = ctx.Value(dataSrcKey).(string)
	return
}
