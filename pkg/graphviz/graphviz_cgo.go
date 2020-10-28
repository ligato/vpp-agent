//  Copyright (c) 2020 Cisco and/or its affiliates.
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

// +build cgo

package graphviz

import (
	"bytes"
	"log"

	"github.com/goccy/go-graphviz"
)

func RenderFilename(outfname, format string, dot []byte) error {
	g, err := graphviz.ParseBytes(dot)
	if err != nil {
		return err
	}

	gv := graphviz.New()
	defer func() {
		if err := g.Close(); err != nil {
			log.Println("dotgraph: closing graph: %w", err)
		}
		_ = gv.Close()
	}()

	err = gv.RenderFilename(g, graphviz.Format(format), outfname)
	if err != nil {
		return err
	}

	return nil
}

func RenderDot(dot []byte) ([]byte, error) {
	g, err := graphviz.ParseBytes(dot)
	if err != nil {
		return nil, err
	}

	gv := graphviz.New()
	defer func() {
		if err := g.Close(); err != nil {
			log.Println("dotgraph: closing graph: %w", err)
		}
		_ = gv.Close()
	}()

	var buf bytes.Buffer
	err = gv.Render(g, "dot", &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
