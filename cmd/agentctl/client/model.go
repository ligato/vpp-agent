package client

import (
	"context"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/ligato/vpp-agent/api/generic"
	"github.com/ligato/vpp-agent/api/types"
	"github.com/ligato/vpp-agent/pkg/debug"
	"github.com/ligato/vpp-agent/pkg/models"
)

func (c *Client) ModelList(ctx context.Context, opts types.ModelListOptions) ([]types.Model, error) {
	cfgClient, err := c.ConfigClient()
	if err != nil {
		return nil, err
	}
	knownModels, err := cfgClient.KnownModels(opts.Class)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("retrieved %d known models", len(knownModels))
	if debug.IsEnabledFor("models") {
		for _, m := range knownModels {
			logrus.Debug(proto.CompactTextString(m))
		}
	}

	allModels := convertModels(knownModels)
	sort.Sort(modelsByName(allModels))

	return allModels, nil
}

func convertModels(knownModels []*generic.ModelDescriptor) []types.Model {
	allModels := make([]types.Model, len(knownModels))
	for i, m := range knownModels {
		spec := models.Spec(*m.Spec)

		protoName := m.ProtoName
		keyPrefix := spec.KeyPrefix()

		var (
			nameTemplate string
			goType       string
		)
		for _, o := range m.Options {
			if o.Key == "nameTemplate" && len(o.Values) > 0 {
				nameTemplate = o.Values[0]
			}
			if o.Key == "goType" && len(o.Values) > 0 {
				goType = o.Values[0]
			}
		}

		model := types.Model{
			Name:         spec.ModelName(),
			Module:       m.Spec.Module,
			Version:      m.Spec.Version,
			Type:         m.Spec.Type,
			Class:        m.Spec.Class,
			KeyPrefix:    keyPrefix,
			ProtoName:    protoName,
			NameTemplate: nameTemplate,
			GoType:       goType,
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
