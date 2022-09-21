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

type labelsKeyT string

var labelsKey = labelsKeyT("labels")

func LabelsContext(ctx context.Context, labels map[string]string) context.Context {
	return context.WithValue(ctx, labelsKey, labels)
}

func LabelsFromContext(ctx context.Context) (labels map[string]string, ok bool) {
	labels, ok = ctx.Value(labelsKey).(map[string]string)
	return
}

// TODO: This is hack to avoid import cycle between orchestrator and contextdecorator package.
// Figure out a way to pass result into local client without using interface implemented by
// a wrapper type defined inside orchestrator package.
type resulter interface {
	IsPushDataResult()
}

type pushDataResultKeyT string

var pushDataResultKey = pushDataResultKeyT("pushDataResult")

func PushDataResultContext(ctx context.Context, pushDataResult resulter) context.Context {
	return context.WithValue(ctx, pushDataResultKey, pushDataResult)
}

func PushDataResultFromContext(ctx context.Context) (pushDataResult resulter, ok bool) {
	pushDataResult, ok = ctx.Value(pushDataResultKey).(resulter)
	return
}