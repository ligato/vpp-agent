package client

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/api/types"
	"github.com/ligato/vpp-agent/pkg/debug"
)

func (c *Client) ModelList(ctx context.Context, opts types.ModelListOptions) ([]types.Model, error) {
	cfgClient, err := c.ConfigClient()
	if err != nil {
		return nil, err
	}
	knownModels, err := cfgClient.KnownModels()
	if err != nil {
		return nil, err
	}

	logrus.Debugf("retrieved %d known models", len(knownModels))
	if debug.IsEnabledFor("models") {
		for _, m := range knownModels {
			logrus.Debug(proto.CompactTextString(&m))
		}
	}

	allModels := convertModels(knownModels)
	sort.Sort(modelsByName(allModels))

	return allModels, nil
}

func convertModels(knownModels []genericmanager.ModelInfo) []types.Model {
	allModels := make([]types.Model, len(knownModels))
	for i, m := range knownModels {
		module := strings.Split(m.Model.Module, ".")
		typ := m.Model.Type
		version := m.Model.Version

		name := fmt.Sprintf("%s.%s", m.Model.Module, typ)
		alias := fmt.Sprintf("%s.%s", module[0], typ)
		if alias == name {
			alias = ""
		}

		protoName := m.Info["protoName"]
		keyPrefix := m.Info["keyPrefix"]
		nameTemplate := m.Info["nameTemplate"]

		//p := reflect.New(proto.MessageType(protoName)).Elem().Interface().(descriptor.Message)
		//fd, _ := descriptor.ForMessage(p)

		model := types.Model{
			Name:         name,
			Module:       m.Model.Module,
			Version:      version,
			Type:         typ,
			Alias:        alias,
			KeyPrefix:    keyPrefix,
			ProtoName:    protoName,
			NameTemplate: nameTemplate,
			//Proto:    proto.MarshalTextString(fd),
			//ProtoFile: fd.GetName(),
		}
		allModels[i] = model
	}
	return allModels
}

type modelsByName []types.Model

func (s modelsByName) Len() int {
	return len(s)
}

func (s modelsByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s modelsByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
