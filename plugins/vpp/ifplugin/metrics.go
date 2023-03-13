package ifplugin

import (
	"github.com/prometheus/client_golang/prometheus"
)

var operationalStates = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ligato",
	Subsystem: "ifplugin",
	Name:      "operational_state",
	Help:      "The operational state of available interfaces.",
}, []string{"name"})
var adminStates = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ligato",
	Subsystem: "ifplugin",
	Name:      "admin_state",
	Help:      "The admin state of available interfaces.",
}, []string{"name"})

func registerMetrics() {
	prometheus.MustRegister(operationalStates)
	prometheus.MustRegister(adminStates)
}
