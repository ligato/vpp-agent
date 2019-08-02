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
	"html/template"
)

type logType struct {
	Logger string `json:"logger,omitempty"`
	Level  string `json:"level,omitempty"`
}

type LogList []logType

func ConvertToLogList(log string) LogList {
	data := make(LogList, 0)
	err := json.Unmarshal([]byte(log), &data)

	if err != nil {
		ExitWithError(ExitError,
			errors.New("Failed conver string to json - "+err.Error()))
	}

	return data
}

func (ll LogList) PrintLogList() (*bytes.Buffer, error) {

	logLevel := createLogTypeTemplate()

	templates := []*template.Template{}

	templates = append(templates, logLevel)

	return ll.textRenderer(templates)
}

func createLogTypeTemplate() *template.Template {
	Template := template.Must(template.New("log").Parse(
		"{{with .Logger}}\nLogger: {{.}}{{end}}" +
			"{{with .Level}}\nLevel: {{.}}{{end}}"))

	return Template
}

func (ll LogList) textRenderer(templates []*template.Template) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	for _, value := range ll {
		for _, templateVal := range templates {
			err := templateVal.Execute(buffer, value)
			if err != nil {
				return nil, err
			}
		}
	}

	return buffer, nil
}
