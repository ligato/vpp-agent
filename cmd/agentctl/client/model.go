package client

import (
	"context"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/pkg/debug"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
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

func convertModels(knownModels []*generic.ModelDetail) []types.Model {
	allModels := make([]types.Model, len(knownModels))
	for i, m := range knownModels {
		spec := models.ToSpec(m.Spec)

		protoName := m.GetProtoName()
		keyPrefix := spec.KeyPrefix()

		var (
			nameTemplate string
			goType       string
		)
		for _, o := range m.Options {
			if o.GetKey() == "nameTemplate" && len(o.Values) > 0 {
				nameTemplate = o.Values[0]
			}
			if o.GetKey() == "goType" && len(o.Values) > 0 {
				goType = o.Values[0]
			}
		}

		model := types.Model{
			Name:         spec.ModelName(),
			Module:       spec.Module,
			Version:      spec.Version,
			Type:         spec.Type,
			Class:        spec.Class,
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
