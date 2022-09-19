package client

import (
	"context"
	"sort"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/pkg/debug"
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

func (c *Client) ModelList(ctx context.Context, opts types.ModelListOptions) ([]types.Model, error) {
	cfgClient, err := c.GenericClient()
	if err != nil {
		return nil, err
	}
	knownModels, err := cfgClient.KnownModels(opts.Class)
	if err != nil {
		return nil, err
	}
	for _, km := range knownModels {
		kmSpec := models.ToSpec(km.GetSpec())
		if _, err = models.DefaultRegistry.GetModel(kmSpec.ModelName()); err != nil {
			if _, err = models.DefaultRegistry.Register(km, kmSpec); err != nil {
				return nil, err
			}
		}
	}
	logrus.Debugf("retrieved %d known models", len(knownModels))
	if debug.IsEnabledFor("models") {
		for _, km := range knownModels {
			logrus.Trace(" - ", prototext.Format(km))
		}
	}
	allModels := convertModels(knownModels)

	return sortModels(allModels), nil
}

// sortModels sorts models in this order:
//	Class > Name > Version
func sortModels(list []types.Model) []types.Model {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Class != list[j].Class {
			return list[i].Class < list[j].Class
		}
		if list[i].Name != list[j].Name {
			return list[i].Name < list[j].Name
		}
		return list[i].Version < list[j].Version
	})
	return list
}

func convertModels(knownModels []*models.ModelInfo) []types.Model {
	allModels := make([]types.Model, len(knownModels))
	for i, m := range knownModels {
		spec := models.ToSpec(m.Spec)

		protoName := m.GetProtoName()
		keyPrefix := spec.KeyPrefix()

		var (
			nameTemplate string
			goType       string
			pkgPath      string
			protoFile    string
		)
		for _, o := range m.Options {
			if o.GetKey() == "nameTemplate" && len(o.Values) > 0 {
				nameTemplate = o.Values[0]
			}
			if o.GetKey() == "goType" && len(o.Values) > 0 {
				goType = o.Values[0]
			}
			if o.GetKey() == "pkgPath" && len(o.Values) > 0 {
				pkgPath = o.Values[0]
			}
			if o.GetKey() == "protoFile" && len(o.Values) > 0 {
				protoFile = o.Values[0]
			}
		}
		// fix key prefixes for models with no template
		if nameTemplate == "" {
			km, err := models.GetModel(spec.ModelName())
			if err == nil && km.KeyPrefix() != keyPrefix {
				logrus.Debugf("key prefix for model %v fixed from %q to %q", spec.ModelName(), keyPrefix, km.KeyPrefix())
				keyPrefix = km.KeyPrefix()
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
			ProtoFile:    protoFile,
			NameTemplate: nameTemplate,
			GoType:       goType,
			PkgPath:      pkgPath,
		}
		allModels[i] = model
	}
	return allModels
}
