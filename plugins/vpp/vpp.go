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

package vpp

import (
	"errors"

	govppapi "git.fd.io/govpp.git/api"
)

var (
	// ErrIncompatible is an error returned when no compatible handler is found.
	ErrIncompatible = errors.New("incompatible handler")

	// ErrNoVersions is an error returned when no handler versions are found.
	ErrNoVersions = errors.New("no handler versions")

	// ErrPluginDisabled is an error returned when disabled plugin is detected.
	ErrPluginDisabled = errors.New("plugin not available")
)

// Version defines a VPP version
type Version string

type APIChannel interface {
	govppapi.Channel
}

// Client provides methods for managing VPP.
type Client interface {
	IsPluginLoaded(plugin string) bool

	govppapi.ChannelProvider

	// Stats provides access to VPP stats API.
	Stats() govppapi.StatsProvider

	CheckCompatiblity(...govppapi.Message) error

	StatsConnected() bool
}
