package mymodel

import (
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

func init() {
	models.Register(&MyModel{}, models.Spec{
		Module:  "custom",
		Type:    "mymodel",
		Version: "v2",
	})
}
