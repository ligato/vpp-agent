package telemetry

import (
	"context"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	debug = os.Getenv("DEBUG_TELEMETRY") != ""
)

const (
	// Registry path for telemetry metrics
	registryPath = "/vpp"

	// Metrics label used for agent label
	agentLabel = "agent"

	// Runtime
	runtimeThreadLabel   = "thread"
	runtimeThreadIDLabel = "threadID"
	runtimeItemLabel     = "item"

	runtimeCallsMetric          = "calls"
	runtimeVectorsMetric        = "vectors"
	runtimeSuspendsMetric       = "suspends"
	runtimeClocksMetric         = "clocks"
	runtimeVectorsPerCallMetric = "vectors_per_call"

	// Memory
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

	// Buffers
	buffersThreadIDLabel = "threadID"
	buffersItemLabel     = "item"
	buffersIndexLabel    = "index"

	buffersSizeMetric     = "size"
	buffersAllocMetric    = "alloc"
	buffersFreeMetric     = "free"
	buffersNumAllocMetric = "num_alloc"
	buffersNumFreeMetric  = "num_free"

	// Node counters
	nodeCounterItemLabel   = "item"
	nodeCounterReasonLabel = "reason"

	nodeCounterCountMetric = "count"
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

func (p *Plugin) registerPrometheus() error {
	p.Log.Debugf("registering prometheus registry path: %v", registryPath)

	// Register '/vpp' registry path
	err := p.Prometheus.NewRegistry(registryPath, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})
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
			Namespace: "vpp",
			Subsystem: "runtime",
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
			Namespace: "vpp",
			Subsystem: "buffers",
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
		{nodeCounterCountMetric, "Count"},
	} {
		name := metric[0]
		p.nodeCounterGaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "vpp",
			Subsystem: "node_counter",
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
		for _, thread := range runtimeInfo.Threads {
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
		for _, item := range buffersInfo.Items {
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

	// Update node counters
	nodeCountersInfo, err := p.handler.GetNodeCounters(ctx)
	if err != nil {
		p.Log.Errorf("GetNodeCounters failed: %v", err)
	} else {
		p.tracef("node counters info: %+v", nodeCountersInfo)
		for _, item := range nodeCountersInfo.Counters {
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

			stats.metrics[nodeCounterCountMetric].Set(float64(item.Value))
		}
	}

	// Update memory
	memoryInfo, err := p.handler.GetMemory(ctx)
	if err != nil {
		p.Log.Errorf("GetMemory failed: %v", err)
	} else {
		p.tracef("memory info: %+v", memoryInfo)
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
			stats.metrics[memorySizeMetric].Set(float64(thread.Size))
			stats.metrics[memoryPagesMetric].Set(float64(thread.Pages))
		}
	}
}
