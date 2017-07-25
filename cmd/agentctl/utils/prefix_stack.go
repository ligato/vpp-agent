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

package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type pfxStackEntry struct {
	preamble string
	last     bool
}

type pfxStack struct {
	entries    []pfxStackEntry
	spaces     int
	firstDash  string
	middleDash string
	lastDash   string
}

func (stack *pfxStack) getPrefix() string {
	var pfx = ""
	for _, se := range stack.entries {
		pfx = pfx + se.preamble
	}
	return pfx
}

// setLast sets the top element of the prefix stack to display
// the last element of a list.
func (stack *pfxStack) setLast() {
	stack.entries[len(stack.entries)-1].preamble = stack.getPreamble(stack.lastDash)
	stack.entries[len(stack.entries)-1].last = true
}

// push increases the current prefix stack level (i.e. it makes the prefix
// stack longer. If the list at the current level continues (i.e. the
// list element is not the last element), the current prefix element
// is replaced with a vertical bar (middleDash) icon. If the current
// element is the last element of a list, the current prefix element
// is replaced with a space (i.e. the vertical line in the tree will
// not continue).
func (stack *pfxStack) push() {
	// Replace current entry at the top of the prefix stack with either
	// vertical bar or empty space.
	if len(stack.entries) > 0 {
		if stack.entries[len(stack.entries)-1].last {
			stack.entries[len(stack.entries)-1].preamble =
				fmt.Sprintf("%s",
					strings.Repeat(" ", stack.spaces+utf8.RuneCountInString(stack.lastDash)))

		} else {
			stack.entries[len(stack.entries)-1].preamble = stack.getPreamble(stack.middleDash)
		}

	}
	// Add new entry at the top of the prefix stack
	stack.entries = append(stack.entries, pfxStackEntry{
		preamble: stack.getPreamble(stack.firstDash),
		last:     false})
}

// pop increases the current prefix stack level (i.e. it make the
// prefix stack shorter. If after pop the element at the top of the
// prefix stack is not the the last element on a list, it's replaced
// by the list element (firstDash) icon.
func (stack *pfxStack) pop() {
	stack.entries = stack.entries[:len(stack.entries)-1]
	if !stack.entries[len(stack.entries)-1].last {
		stack.entries[len(stack.entries)-1].preamble = stack.getPreamble(stack.firstDash)
	}
}

// getPreamble creates the string for a prefix stack entry. The
// prefix itself is then created by joining all prefix entries.
func (stack *pfxStack) getPreamble(icon string) string {
	return fmt.Sprintf("%s%s", strings.Repeat(" ", stack.spaces), icon)
}

func (stack *pfxStack) setTopPfxStackEntry(new string) string {
	prev := stack.entries[len(stack.entries)-1].preamble
	stack.entries[len(stack.entries)-1].preamble = new
	return prev
}

func (stack *pfxStack) getTopPfxStackEntry() string {
	return stack.entries[len(stack.entries)-1].preamble
}
