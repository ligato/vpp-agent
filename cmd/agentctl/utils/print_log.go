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

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	_ "sort"
	"text/tabwriter"
	"text/template"
)

type logType struct {
	Logger string `json:"logger,omitempty"`
	Level  string `json:"level,omitempty"`
}

type LogList []logType

func ConvertToLogList(log string) (LogList, error) {
	data := make(LogList, 0)
	err := json.Unmarshal([]byte(log), &data)
	if err != nil {
		return nil, errors.New("Failed conver string to json - " + err.Error())
	}
	sort.Sort(data)
	return data, nil
}

func (ll LogList) Print(w io.Writer) error {
	const tmpl = "{{.Logger}}\t{{.Level}}\t\n"
	t, err := template.New("log").Parse(tmpl)
	if err != nil {
		return err
	}
	b, err := ll.render(t)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (ll LogList) render(t *template.Template) ([]byte, error) {
	var buffer bytes.Buffer
	w := tabwriter.NewWriter(&buffer, 0, 0, 1, ' ', 0)

	// print header
	fmt.Fprintf(w, "LOGGER\tLEVEL\t\n")

	// print logger list
	for _, value := range ll {
		err := t.Execute(w, value)
		if err != nil {
			return nil, err
		}
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (ll LogList) Len() int {
	return len(ll)
}

func (ll LogList) Less(i, j int) bool {
	return ll[i].Logger < ll[j].Logger
}

func (ll LogList) Swap(i, j int) {
	ll[i], ll[j] = ll[j], ll[i]
}
