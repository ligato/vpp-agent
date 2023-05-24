//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package commands

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

func protoMessageType(fullName string) protoreflect.MessageType {
	valueType, err := models.DefaultRegistry.MessageTypeRegistry().FindMessageByName(protoreflect.FullName(fullName))
	if err != nil {
		logrus.Errorf("error finding message with name %q: %v", fullName, err)
		return nil
	}
	return valueType
}
