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
	"fmt"
	"sort"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
)

// HandlerDesc is a handler descriptor used to describe a handler for registration.
type HandlerDesc struct {
	Name       string
	HandlerAPI interface{}
	NewFunc    interface{}
}

// HandlerVersion defines handler implementation for specific version used by AddVersion.
type HandlerVersion struct {
	Version    string
	Check      func(Client) error
	NewHandler func(Client, ...interface{}) HandlerAPI
}

// HandlerAPI is an empty interface representing handler interface.
type HandlerAPI interface{}

// Handler is a handler for managing implementations for multiple versions.
type Handler struct {
	desc     *HandlerDesc
	versions map[string]*HandlerVersion
}

// AddVersion adds handler version to a list of available versions.
// Handler versions can be overwritten by calling AddVersion multiple times.
func (h *Handler) AddVersion(hv HandlerVersion) {
	if _, ok := h.versions[hv.Version]; ok {
		logging.Warnf("overwritting %s handler version: %s", h.desc.Name, hv.Version)
	}
	// TODO: check if given handler version implementes handler API interface
	/*ht := reflect.TypeOf(h.desc.HandlerAPI).Elem()
	hc := reflect.TypeOf(hv.New).Out(0)
	if !hc.Implements(ht) {
		logging.DefaultLogger.Warnf("vpphandlers: AddVersion found the handler of type %v that does not satisfy %v", hc, ht)
	}*/
	h.versions[hv.Version] = &hv
}

// Versions returns list of versions from list of available handler versions.
func (h *Handler) Versions() []string {
	var versions []string
	for _, v := range h.versions {
		versions = append(versions, v.Version)
	}
	sort.Strings(versions)
	return versions
}

// FindCompatibleVersion iterates over all available handler versions and calls
// their Check method to check compatibility.
func (h *Handler) FindCompatibleVersion(c Client) *HandlerVersion {
	v, err := h.GetCompatibleVersion(c)
	if err != nil {
		return nil
	}
	return v
}

// GetCompatibleVersion iterates over all available handler versions and calls
// // their Check method to check compatibility.
func (h *Handler) GetCompatibleVersion(c Client) (*HandlerVersion, error) {
	if len(h.versions) == 0 {
		logging.Debugf("VPP handler %s has no registered versions", h.desc.Name)
		return nil, ErrNoVersions
	}
	for _, v := range h.versions {
		if err := v.Check(c); err != nil {
			if ierr, ok := err.(*govppapi.CompatibilityError); ok {
				logging.Debugf("VPP handler %s incompatible with version %s: found %d incompatible messages",
					h.desc.Name, v.Version, len(ierr.IncompatibleMessages))
			} else {
				logging.Debugf("VPP handler %s failed check for version %s: \n%v",
					h.desc.Name, v.Version, err)
			}
			continue
		}
		logging.Debugf("VPP handler %s compatible version: %s", h.desc.Name, v.Version)
		return v, nil
	}
	return nil, ErrIncompatible
}

var (
	handlers = map[string]*Handler{}
)

// RegisterHandler creates new handler described by handle descriptor.
func RegisterHandler(hd HandlerDesc) *Handler {
	if _, ok := handlers[hd.Name]; ok {
		panic(fmt.Sprintf("VPP handler %s is already registered", hd.Name))
	}
	h := &Handler{
		desc:     &hd,
		versions: make(map[string]*HandlerVersion),
	}
	handlers[hd.Name] = h
	return h
}

// GetHandlers returns map for all registered handlers.
func GetHandlers() map[string]*Handler {
	return handlers
}
