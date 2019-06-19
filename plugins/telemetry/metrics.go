package telemetry

import (
	"context"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Registry path for telemetry metrics
	registryPath = "/metrics/vpp"

	vppMetricsNamespace = "vpp"

	// Metrics label used for agent label
	agentLabel = "agent"
)

// Runtime metrics
const (
	runtimeMetricsNamespace = "runtime"

	runtimeThreadLabel   = "thread"
	runtimeThreadIDLabel = "threadID"
	runtimeItemLabel     = "item"

	runtimeCallsMetric          = "calls"
	runtimeVectorsMetric        = "vectors"
	runtimeSuspendsMetric       = "suspends"
	runtimeClocksMetric         = "clocks"
	runtimeVectorsPerCallMetric = "vectors_per_call"
)

// Memory metrics
const (
	memoryMetricsNamespace = "memory"

	memoryThreadLabel   = "thread"
	memoryThreadIDLabel = "threadID"

	memoryObjectsMetric   = "objects"
	memoryUsedMetric      = "used"
	memoryTotalMetric     = "total"
	memoryFreeMetric      = "free"
	memoryReclaimedMetric = "reclaimed"
	memoryOverheadMetric  = "overhead"
	memorySizeMetric      = "size"
	memoryPagesMetric     = "pages"
)

// Buffers metrics
const (
	buffersMetricsNamespace = "buffers"

	buffersThreadIDLabel = "threadID"
	buffersItemLabel     = "item"
	buffersIndexLabel    = "index"

	buffersSizeMetric     = "size"
	buffersAllocMetric    = "alloc"
	buffersFreeMetric     = "free"
	buffersNumAllocMetric = "num_alloc"
	buffersNumFreeMetric  = "num_free"
)

// Node metrics
const (
	nodeMetricsNamespace = "nodes"

	nodeCounterItemLabel   = "item"
	nodeCounterReasonLabel = "reason"

	nodeCounterCounterMetric = "counter"
)

// Interface metrics
const (
	ifMetricsNamespace = "interfaces"

	ifCounterNameLabel  = "name"
	ifCounterIndexLabel = "index"

	ifCounterRxPackets = "rx_packets"
	ifCounterRxBytes   = "rx_bytes"
	ifCounterRxErrors  = "rx_errors"
	ifCounterTxPackets = "tx_packets"
	ifCounterTxBytes   = "tx_bytes"
	ifCounterTxErrors  = "tx_errors"
	ifCounterDrops     = "drops"
	ifCounterPunts     = "punts"
	ifCounterIP4       = "ip4"
	ifCounterIP6       = "ip6"
	ifCounterRxNoBuf   = "rx_no_buf"
	ifCounterRxMiss    = "rx_miss"
)

type prometheusMetrics struct {
	runtimeGaugeVecs map[string]*prometheus.GaugeVec
	runtimeStats     map[string]*runtimeStats

	memoryGaugeVecs map[string]*prometheus.GaugeVec
	memoryStats     map[string]*memoryStats

	buffersGaugeVecs map[string]*prometheus.GaugeVec
	buffersStats     map[string]*buffersStats

	nodeCounterGaugeVecs map[string]*prometheus.GaugeVec
	nodeCounterStats     map[string]*nodeCounterStats

	ifCounterGaugeVecs map[string]*prometheus.GaugeVec
	ifCounterStats     map[string]*ifCounterStats
}

type runtimeStats struct {
	threadName string
	threadID   uint
	itemName   string
	metrics    map[string]prometheus.Gauge
}

type memoryStats struct {
	threadName string
	threadID   uint
	metrics    map[string]prometheus.Gauge
}

type buffersStats struct {
	threadID  uint
	itemName  string
	itemIndex uint
	metrics   map[string]prometheus.Gauge
}

type nodeCounterStats struct {
	itemName string
	metrics  map[string]prometheus.Gauge
}

type ifCounterStats struct {
	name    string
	metrics map[string]prometheus.Gauge
}

func (p *Plugin) registerPrometheus() error {
	p.Log.Debugf("registering prometheus registry path: %v", registryPath)

	// Register vpp registry path
	err := p.Prometheus.NewRegistry(registryPath, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	})
	if err != nil {
		return err
	}

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
			Namespace: vppMetricsNamespace,
			Subsystem: runtimeMetricsNamespace,
			Name:      name,
			Help:      metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: p.ServiceLabel.GetAgentLabel(),
			},
		}, []string{runtimeItemLabel, runtimeThreadLabel, runtimeThreadIDLabel})

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
		{memorySizeMetric, "Size"},
		{memoryPagesMetric, "Pages"},
	} {
		name := metric[0]
		p.memoryGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: vppMetricsNamespace,
			Subsystem: memoryMetricsNamespace,
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

	// Buffers metrics
	p.buffersGaugeVecs = make(map[string]*prometheus.GaugeVec)
	p.buffersStats = make(map[string]*buffersStats)

	for _, metric := range [][2]string{
		{buffersSizeMetric, "Size of buffer"},
		{buffersAllocMetric, "Allocated"},
		{buffersFreeMetric, "Free"},
		{buffersNumAllocMetric, "Number of allocated"},
		{buffersNumFreeMetric, "Number of free"},
	} {
		name := metric[0]
		p.buffersGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: vppMetricsNamespace,
			Subsystem: buffersMetricsNamespace,
			Name:      name,
			Help:      metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: p.ServiceLabel.GetAgentLabel(),
			},
		}, []string{buffersThreadIDLabel, buffersItemLabel, buffersIndexLabel})

	}

	// register created vectors to prometheus
	for name, metric := range p.buffersGaugeVecs {
		if err := p.Prometheus.Register(registryPath, metric); err != nil {
			p.Log.Errorf("failed to register %v metric: %v", name, err)
			return err
		}
	}

	// Node counters metrics
	p.nodeCounterGaugeVecs = make(map[string]*prometheus.GaugeVec)
	p.nodeCounterStats = make(map[string]*nodeCounterStats)

	for _, metric := range [][2]string{
		{nodeCounterCounterMetric, "Counter"},
	} {
		name := metric[0]
		p.nodeCounterGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: vppMetricsNamespace,
			Subsystem: nodeMetricsNamespace,
			Name:      name,
			Help:      metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: p.ServiceLabel.GetAgentLabel(),
			},
		}, []string{nodeCounterItemLabel, nodeCounterReasonLabel})

	}

	// register created vectors to prometheus
	for name, metric := range p.nodeCounterGaugeVecs {
		if err := p.Prometheus.Register(registryPath, metric); err != nil {
			p.Log.Errorf("failed to register %v metric: %v", name, err)
			return err
		}
	}

	// Interface counter metrics
	p.ifCounterGaugeVecs = make(map[string]*prometheus.GaugeVec)
	p.ifCounterStats = make(map[string]*ifCounterStats)

	for _, metric := range [][2]string{
		{ifCounterRxPackets, "RX packets"},
		{ifCounterRxBytes, "RX bytes"},
		{ifCounterRxErrors, "RX errors"},
		{ifCounterTxPackets, "TX packets"},
		{ifCounterTxBytes, "TX bytes"},
		{ifCounterTxErrors, "TX errors"},
		{ifCounterDrops, "Drops"},
		{ifCounterPunts, "Punts"},
		{ifCounterIP4, "IP4"},
		{ifCounterIP6, "IP6"},
		{ifCounterRxNoBuf, "RX nobuf"},
		{ifCounterRxMiss, "RX miss"},
	} {
		name := metric[0]
		p.ifCounterGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: vppMetricsNamespace,
			Subsystem: ifMetricsNamespace,
			Name:      name,
			Help:      metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: p.ServiceLabel.GetAgentLabel(),
			},
		}, []string{ifCounterNameLabel, ifCounterIndexLabel})

	}

	// register created vectors to prometheus
	for name, metric := range p.ifCounterGaugeVecs {
		if err := p.Prometheus.Register(registryPath, metric); err != nil {
			p.Log.Errorf("failed to register %v metric: %v", name, err)
			return err
		}
	}

	return nil
}

func (p *Plugin) updatePrometheus(ctx context.Context) {
	p.tracef("running update")

	// Update runtime
	runtimeInfo, err := p.handler.GetRuntimeInfo(ctx)
	if err != nil {
		p.Log.Errorf("GetRuntimeInfo failed: %v", err)
	} else {
		p.tracef("runtime info: %+v", runtimeInfo)
		for _, thread := range runtimeInfo.GetThreads() {
			for _, item := range thread.Items {
				stats, ok := p.runtimeStats[item.Name]
				if !ok {
					stats = &runtimeStats{
						threadID:   thread.ID,
						threadName: thread.Name,
						itemName:   item.Name,
						metrics:    map[string]prometheus.Gauge{},
					}

					// add gauges with corresponding labels into vectors
					for k, vec := range p.runtimeGaugeVecs {
						stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
							runtimeItemLabel:     item.Name,
							runtimeThreadLabel:   thread.Name,
							runtimeThreadIDLabel: strconv.Itoa(int(thread.ID)),
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
				stats.metrics[runtimeVectorsPerCallMetric].Set(item.VectorsPerCall)
			}
		}
	}

	// Update buffers
	buffersInfo, err := p.handler.GetBuffersInfo(ctx)
	if err != nil {
		p.Log.Errorf("GetBuffersInfo failed: %v", err)
	} else {
		p.tracef("buffers info: %+v", buffersInfo)
		for _, item := range buffersInfo.GetItems() {
			stats, ok := p.buffersStats[item.Name]
			if !ok {
				stats = &buffersStats{
					threadID:  item.ThreadID,
					itemName:  item.Name,
					itemIndex: item.Index,
					metrics:   map[string]prometheus.Gauge{},
				}

				// add gauges with corresponding labels into vectors
				for k, vec := range p.buffersGaugeVecs {
					stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
						buffersThreadIDLabel: strconv.Itoa(int(item.ThreadID)),
						buffersItemLabel:     item.Name,
						buffersIndexLabel:    strconv.Itoa(int(item.Index)),
					})
					if err != nil {
						p.Log.Error(err)
					}
				}
			}

			stats.metrics[buffersSizeMetric].Set(float64(item.Size))
			stats.metrics[buffersAllocMetric].Set(float64(item.Alloc))
			stats.metrics[buffersFreeMetric].Set(float64(item.Free))
			stats.metrics[buffersNumAllocMetric].Set(float64(item.NumAlloc))
			stats.metrics[buffersNumFreeMetric].Set(float64(item.NumFree))
		}
	}

	// Update memory
	memoryInfo, err := p.handler.GetMemory(ctx)
	if err != nil {
		p.Log.Errorf("GetMemory failed: %v", err)
	} else {
		p.tracef("memory info: %+v", memoryInfo)
		for _, thread := range memoryInfo.GetThreads() {
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
			stats.metrics[memorySizeMetric].Set(float64(thread.Size))
			stats.metrics[memoryPagesMetric].Set(float64(thread.Pages))
		}
	}

	// Update node counters
	nodeCountersInfo, err := p.handler.GetNodeCounters(ctx)
	if err != nil {
		p.Log.Errorf("GetNodeCounters failed: %v", err)
	} else {
		p.tracef("node counters info: %+v", nodeCountersInfo)
		for _, item := range nodeCountersInfo.GetCounters() {
			stats, ok := p.nodeCounterStats[item.Name]
			if !ok {
				stats = &nodeCounterStats{
					itemName: item.Name,
					metrics:  map[string]prometheus.Gauge{},
				}

				// add gauges with corresponding labels into vectors
				for k, vec := range p.nodeCounterGaugeVecs {
					stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
						nodeCounterItemLabel:   item.Node,
						nodeCounterReasonLabel: item.Name,
					})
					if err != nil {
						p.Log.Error(err)
					}
				}
			}

			stats.metrics[nodeCounterCounterMetric].Set(float64(item.Value))
		}
	}

	// Update interface counters
	ifStats, err := p.handler.GetInterfaceStats(ctx)
	if err != nil {
		p.Log.Errorf("GetInterfaceStats failed: %v", err)
		return
	} else {
		p.tracef("interface stats: %+v", ifStats)
		if ifStats == nil {
			return
		}
		for _, item := range ifStats.Interfaces {
			stats, ok := p.ifCounterStats[item.InterfaceName]
			if !ok {
				stats = &ifCounterStats{
					name:    item.InterfaceName,
					metrics: map[string]prometheus.Gauge{},
				}

				// add gauges with corresponding labels into vectors
				for k, vec := range p.ifCounterGaugeVecs {
					stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
						ifCounterNameLabel:  item.InterfaceName,
						ifCounterIndexLabel: fmt.Sprint(item.InterfaceIndex),
					})
					if err != nil {
						p.Log.Error(err)
					}
				}
			}

			stats.metrics[ifCounterRxPackets].Set(float64(item.RxPackets))
			stats.metrics[ifCounterRxBytes].Set(float64(item.RxBytes))
			stats.metrics[ifCounterRxErrors].Set(float64(item.RxErrors))
			stats.metrics[ifCounterTxPackets].Set(float64(item.TxPackets))
			stats.metrics[ifCounterTxBytes].Set(float64(item.TxBytes))
			stats.metrics[ifCounterTxErrors].Set(float64(item.TxErrors))
			stats.metrics[ifCounterDrops].Set(float64(item.Drops))
			stats.metrics[ifCounterPunts].Set(float64(item.Punts))
			stats.metrics[ifCounterIP4].Set(float64(item.IP4))
			stats.metrics[ifCounterIP6].Set(float64(item.IP6))
			stats.metrics[ifCounterRxNoBuf].Set(float64(item.RxNoBuf))
			stats.metrics[ifCounterRxMiss].Set(float64(item.RxMiss))
		}
	}

	p.tracef("update complete")
}
