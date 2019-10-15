package vpp_client

import (
	"github.com/ligato/vpp-agent/pkg/models"
)

var MetricsModel = models.Register(&Metrics{}, models.Spec{
	Module: "govppmux",
	Type:   "stats",
	Class:  "metrics",
})
