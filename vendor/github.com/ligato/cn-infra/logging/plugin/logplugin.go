// Copyright (c) 2017 Cisco and/or its affiliates.
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

package plugin

import "github.com/ligato/cn-infra/logging"

// Skeleton contains the common parts of logging plugin.
type Skeleton struct {
	factory func(string) (logging.Logger, error)
	reg     func() logging.Registry
}

// NewSkeleton creates new instance of logging skeleton.
func NewSkeleton(factory func(string) (logging.Logger, error), reg func() logging.Registry) *Skeleton {
	return &Skeleton{factory: factory, reg: reg}
}

// Init is called at plugin initialization.
func (lp *Skeleton) Init() error {
	return nil
}

// Close is called at plugin cleanup phase.
func (lp *Skeleton) Close() error {
	return nil
}

// NewLogger creates a new instance of named logger.
func (lp *Skeleton) NewLogger(name string) (logging.Logger, error) {
	return lp.factory(name)
}

// Registry returns logger registry that can be used for management of loggers.
func (lp *Skeleton) Registry() logging.Registry {
	return lp.reg()
}
