package telemetry

import (
	"strconv"
	"time"

	"github.com/ligato/cn-infra/flavors/local"
	prom "github.com/ligato/cn-infra/rpc/prometheus"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	registryPath = prom.DefaultRegistry

	agentLabel = "agent"

	// Runtime metrics
	runtimeItemLabel = "item"

	runtimeCallsMetric          = "calls"
	runtimeVectorsMetric        = "vectors"
	runtimeSuspendsMetric       = "suspends"
	runtimeClocksMetric         = "clocks"
	runtimeVectorsPerCallMetric = "vectorsPerCall"

	// Memory metrics
	memoryThreadLabel   = "thread"
	memoryThreadIDLabel = "threadid"

	memoryObjectsMetric   = "objects"
	memoryUsedMetric      = "used"
	memoryTotalMetric     = "total"
	memoryFreeMetric      = "free"
	memoryReclaimedMetric = "reclaimed"
	memoryOverheadMetric  = "overhead"
	memoryCapacityMetric  = "capacity"
)

// Plugin registers Telemetry Plugin
type Plugin struct {
	Deps

	runtimeGaugeVecs map[string]*prometheus.GaugeVec
	runtimeStats     map[string]*runtimeStats

	memoryGaugeVecs map[string]*prometheus.GaugeVec
	memoryStats     map[string]*memoryStats
}

// Deps represents dependencies of Telemetry Plugin
type Deps struct {
	local.PluginInfraDeps

	GoVppmux   govppmux.API
	Prometheus prom.API
}

type runtimeStats struct {
	itemName string
	metrics  map[string]prometheus.Gauge
}

type memoryStats struct {
	threadName string
	threadID   uint
	metrics    map[string]prometheus.Gauge
}

// Init initializes Telemetry Plugin
func (p *Plugin) Init() error {
	// Runtime metrics
	p.runtimeGaugeVecs = make(map[string]*prometheus.GaugeVec)
	p.runtimeStats = make(map[string]*runtimeStats)

	for _, metric := range [][2]string{
		{runtimeCallsMetric, "Number of calls"},
		{runtimeVectorsMetric, "Number of vectors"},
		{runtimeSuspendsMetric, "Number of suspends"},
		{runtimeClocksMetric, "Number of clocks"},
		{runtimeVectorsPerCallMetric, "Number of vectors per call"},
	} {
		name := metric[0]
		p.runtimeGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "vpp",
			Subsystem: "runtime",
			Name:      name,
			Help:      metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: p.ServiceLabel.GetAgentLabel(),
			},
		}, []string{runtimeItemLabel})

	}

	// register created vectors to prometheus
	for name, metric := range p.runtimeGaugeVecs {
		if err := p.Prometheus.Register(registryPath, metric); err != nil {
			p.Log.Errorf("failed to register %v metric: %v", name, err)
			return err
		}
	}

	// Memory metrics
	p.memoryGaugeVecs = make(map[string]*prometheus.GaugeVec)
	p.memoryStats = make(map[string]*memoryStats)

	for _, metric := range [][2]string{
		{memoryObjectsMetric, "Number of objects"},
		{memoryUsedMetric, "Used memory"},
		{memoryTotalMetric, "Total memory"},
		{memoryFreeMetric, "Free memory"},
		{memoryReclaimedMetric, "Reclaimed memory"},
		{memoryOverheadMetric, "Overhead"},
		{memoryCapacityMetric, "Capacity"},
	} {
		name := metric[0]
		p.memoryGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "vpp",
			Subsystem: "memory",
			Name:      name,
			Help:      metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: p.ServiceLabel.GetAgentLabel(),
			},
		}, []string{memoryThreadLabel, memoryThreadIDLabel})

	}

	// register created vectors to prometheus
	for name, metric := range p.memoryGaugeVecs {
		if err := p.Prometheus.Register(registryPath, metric); err != nil {
			p.Log.Errorf("failed to register %v metric: %v", name, err)
			return err
		}
	}

	// Update data
	ch, err := p.GoVppmux.NewAPIChannel()
	if err != nil {
		p.Log.Errorf("Error creating channel: %v", err)
		return err
	}

	go func() {
		defer ch.Close()
		for {
			// Update runtime
			runtimeInfo, err := vppcalls.GetRuntimeInfo(ch)
			if err != nil {
				p.Log.Errorf("Sending command failed: %v", err)
				return
			}

			for _, item := range runtimeInfo.Items {
				stats, ok := p.runtimeStats[item.Name]
				if !ok {
					stats = &runtimeStats{
						itemName: item.Name,
						metrics:  map[string]prometheus.Gauge{},
					}

					// add gauges with corresponding labels into vectors
					for k, vec := range p.runtimeGaugeVecs {
						stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
							runtimeItemLabel: item.Name,
						})
						if err != nil {
							p.Log.Error(err)
						}
					}
				}

				stats.metrics[runtimeCallsMetric].Set(float64(item.Calls))
				stats.metrics[runtimeVectorsMetric].Set(float64(item.Vectors))
				stats.metrics[runtimeSuspendsMetric].Set(float64(item.Suspends))
				stats.metrics[runtimeClocksMetric].Set(item.Clocks)
				stats.metrics[runtimeVectorsPerCallMetric].Set(item.VectorsCall)
			}

			// Update memory
			memoryInfo, err := vppcalls.GetMemory(ch)
			if err != nil {
				p.Log.Errorf("Sending command failed: %v", err)
				return
			}

			for _, thread := range memoryInfo.Threads {
				stats, ok := p.memoryStats[thread.Name]
				if !ok {
					stats = &memoryStats{
						threadName: thread.Name,
						threadID:   thread.ID,
						metrics:    map[string]prometheus.Gauge{},
					}

					// add gauges with corresponding labels into vectors
					for k, vec := range p.memoryGaugeVecs {
						stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
							memoryThreadLabel:   thread.Name,
							memoryThreadIDLabel: strconv.Itoa(int(thread.ID)),
						})
						if err != nil {
							p.Log.Error(err)
						}
					}
				}

				stats.metrics[memoryObjectsMetric].Set(float64(thread.Objects))
				stats.metrics[memoryUsedMetric].Set(float64(thread.Used))
				stats.metrics[memoryTotalMetric].Set(float64(thread.Total))
				stats.metrics[memoryFreeMetric].Set(float64(thread.Free))
				stats.metrics[memoryReclaimedMetric].Set(float64(thread.Reclaimed))
				stats.metrics[memoryOverheadMetric].Set(float64(thread.Overhead))
				stats.metrics[memoryCapacityMetric].Set(float64(thread.Capacity))
			}
			time.Sleep(time.Second * 5)
		}
	}()
	return nil
}

// AfterInit executes after initializion of Telemetry Plugin
func (p *Plugin) AfterInit() error {

	return nil
}

// Close is used to clean up resources used by Telemetry Plugin
func (p *Plugin) Close() error {
	return nil
}
