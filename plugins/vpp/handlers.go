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
	"fmt"
	"sort"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"
)

// HandlerVersion defines handler implementation for specific version used by AddVersion.
type HandlerVersion struct {
	Version    Version
	Check      func(Client) error
	NewHandler func(Client, ...interface{}) HandlerAPI
	New        interface{}
}

// HandlerAPI is an empty interface representing handler interface.
type HandlerAPI interface{}

// Handler is a handler for managing implementations for multiple versions.
type Handler struct {
	desc     *HandlerDesc
	versions map[Version]*HandlerVersion
}

func (h *Handler) Name() string {
	return h.desc.Name
}

// AddVersion adds handler version to a list of available versions.
// Handler versions can be overwritten by calling AddVersion multiple times.
func (h *Handler) AddVersion(hv HandlerVersion) {
	if _, ok := h.versions[hv.Version]; ok {
		logging.Warnf("overwritting %s handler version: %s", h.desc.Name, hv.Version)
	}
	if hv.Check == nil {
		panic(fmt.Sprintf("Check not defined for %s handler version: %s", h.desc.Name, hv.Version))
	}
	// TODO: check if given handler version implementes handler API interface
	/*ht := reflect.TypeOf(h.desc.HandlerAPI).Elem()
	  hc := reflect.TypeOf(hv.New).Out(0)
	  if !hc.Implements(ht) {
	  	logging.DefaultLogger.Warnf("vpphandlers: AddVersion found the handler of type %v that does not satisfy %v", hc, ht)
	  }*/
	h.versions[hv.Version] = &hv
}

// FindCompatibleVersion iterates over all available handler versions and calls
// their Check method to check compatibility.
func (h *Handler) FindCompatibleVersion(c Client) *HandlerVersion {
	v, err := h.GetCompatibleVersion(c)
	if err != nil {
		logging.Debugf("no compatible version found for handler %v: %v", h.Name(), err)
		return nil
	}
	logging.Debugf("found compatible version for handler %v: %v", h.Name(), v.Version)
	return v
}

// GetCompatibleVersion iterates over all available handler versions and calls
// their Check method to check compatibility.
func (h *Handler) GetCompatibleVersion(c Client) (*HandlerVersion, error) {
	if len(h.versions) == 0 {
		logging.Debugf("VPP handler %s has no registered versions", h.desc.Name)
		return nil, ErrNoVersions
	}
	// try preferred binapi version first
	if ver := c.BinapiVersion(); ver != "" {
		if v, ok := h.versions[ver]; ok {
			logging.Debugf("VPP handler %s using preferred version: %s", h.desc.Name, v.Version)
			return v, nil
		}
	}
	// fallback to checking all registered versions
	for _, v := range h.versions {
		var compErr *govppapi.CompatibilityError
		if err := v.Check(c); errors.As(err, &compErr) {
			logging.Debugf("VPP handler %s incompatible with %s (%d messages)", h.desc.Name, v.Version, len(compErr.IncompatibleMessages))
		} else if err != nil {
			logging.Warnf("VPP handler %s version %s check failed: %v", h.desc.Name, v.Version, err)
		} else {
			logging.Debugf("VPP handler %s COMPATIBLE with version: %s", h.desc.Name, v.Version)
			return v, nil
		}
	}
	return nil, ErrIncompatible
}

// Versions returns list of versions from list of available handler versions.
func (h *Handler) Versions() []Version {
	vs := make([]Version, 0, len(h.versions))
	for _, v := range h.versions {
		vs = append(vs, v.Version)
	}
	sort.Sort(versions(vs))
	return vs
}

type versions []Version

func (v versions) Len() int           { return len(v) }
func (v versions) Less(i, j int) bool { return v[i] < v[j] }
func (v versions) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

var (
	registeredHandlers = map[string]*Handler{}
)

// HandlerDesc represents a VPP handler's specification.
type HandlerDesc struct {
	Name       string
	HandlerAPI interface{}
	NewFunc    interface{}
}

// RegisterHandler creates new handler described by handle descriptor.
func RegisterHandler(hd HandlerDesc) *Handler {
	if _, ok := registeredHandlers[hd.Name]; ok {
		panic(fmt.Sprintf("VPP handler %s is already registered", hd.Name))
	}
	h := &Handler{
		desc:     &hd,
		versions: make(map[Version]*HandlerVersion),
	}
	registeredHandlers[hd.Name] = h
	return h
}

// GetHandlers returns map for all registered handlers.
func GetHandlers() map[string]*Handler {
	return registeredHandlers
}
