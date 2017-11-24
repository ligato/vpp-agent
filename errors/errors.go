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

package errors

//SwIndexNotFound is specific error type used to differentiate state when software index associated with name
// wasn't found in register
type SwIndexNotFound struct {
	error
	OriginalError error
}

func (swIndexNotFound SwIndexNotFound) Error() string {
	return swIndexNotFound.OriginalError.Error()
}
