package mymodel

import (
	"github.com/ligato/vpp-agent/pkg/models"
)

func init() {
	models.Register(&MyModel{}, models.Spec{
		Module:  "custom",
		Type:    "mymodel",
		Version: "v2",
	})
}
