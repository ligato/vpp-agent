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

package e2e

import "strings"

// parseVPPTable parses table returned by one of the VPP show commands.
func parseVPPTable(table string) (parsed []map[string]string) {
	lines := strings.Split(table, "\r\n")
	if len(lines) == 0 {
		return
	}
	head := lines[0]
	rows := lines[1:]

	var columns []string
	for _, column := range strings.Split(head, " ") {
		if column != "" {
			columns = append(columns, column)
		}
	}
	for _, row := range rows {
		parsedRow := make(map[string]string)
		i := 0
		for _, cell := range strings.Split(row, " ") {
			if cell == "" {
				continue
			}
			if i >= len(columns) {
				break
			}
			parsedRow[columns[i]] = cell
			i++
		}
		if len(parsedRow) > 0 {
			parsed = append(parsed, parsedRow)
		}
	}
	return
}
