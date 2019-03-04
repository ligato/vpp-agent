package core

import (
	"strings"
	"sync/atomic"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/api"
)

const (
	CounterStatsPrefix = "/err/"

	SystemStatsPrefix          = "/sys/"
	SystemStats_VectorRate     = SystemStatsPrefix + "vector_rate"
	SystemStats_InputRate      = SystemStatsPrefix + "input_rate"
	SystemStats_LastUpdate     = SystemStatsPrefix + "last_update"
	SystemStats_LastStatsClear = SystemStatsPrefix + "last_stats_clear"
	SystemStats_Heartbeat      = SystemStatsPrefix + "heartbeat"

	NodeStatsPrefix    = "/sys/node/"
	NodeStats_Clocks   = NodeStatsPrefix + "clocks"
	NodeStats_Vectors  = NodeStatsPrefix + "vectors"
	NodeStats_Calls    = NodeStatsPrefix + "calls"
	NodeStats_Suspends = NodeStatsPrefix + "suspends"

	InterfaceStatsPrefix         = "/if/"
	InterfaceStats_Drops         = InterfaceStatsPrefix + "drops"
	InterfaceStats_Punt          = InterfaceStatsPrefix + "punt"
	InterfaceStats_IP4           = InterfaceStatsPrefix + "ip4"
	InterfaceStats_IP6           = InterfaceStatsPrefix + "ip6"
	InterfaceStats_RxNoBuf       = InterfaceStatsPrefix + "rx-no-buf"
	InterfaceStats_RxMiss        = InterfaceStatsPrefix + "rx-miss"
	InterfaceStats_RxError       = InterfaceStatsPrefix + "rx-error"
	InterfaceStats_TxError       = InterfaceStatsPrefix + "tx-error"
	InterfaceStats_Rx            = InterfaceStatsPrefix + "rx"
	InterfaceStats_RxUnicast     = InterfaceStatsPrefix + "rx-unicast"
	InterfaceStats_RxMulticast   = InterfaceStatsPrefix + "rx-multicast"
	InterfaceStats_RxBroadcast   = InterfaceStatsPrefix + "rx-broadcast"
	InterfaceStats_Tx            = InterfaceStatsPrefix + "tx"
	InterfaceStats_TxUnicastMiss = InterfaceStatsPrefix + "tx-unicast-miss"
	InterfaceStats_TxMulticast   = InterfaceStatsPrefix + "tx-multicast"
	InterfaceStats_TxBroadcast   = InterfaceStatsPrefix + "tx-broadcast"

	NetworkStatsPrefix     = "/net/"
	NetworkStats_RouteTo   = NetworkStatsPrefix + "route/to"
	NetworkStats_RouteVia  = NetworkStatsPrefix + "route/via"
	NetworkStats_MRoute    = NetworkStatsPrefix + "mroute"
	NetworkStats_Adjacency = NetworkStatsPrefix + "adjacency"
)

type StatsConnection struct {
	statsClient adapter.StatsAPI

	connected uint32 // non-zero if the adapter is connected to VPP
}

func newStatsConnection(stats adapter.StatsAPI) *StatsConnection {
	return &StatsConnection{
		statsClient: stats,
	}
}

// Connect connects to Stats API using specified adapter and returns a connection handle.
// This call blocks until it is either connected, or an error occurs.
// Only one connection attempt will be performed.
func ConnectStats(stats adapter.StatsAPI) (*StatsConnection, error) {
	c := newStatsConnection(stats)

	if err := c.connectClient(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *StatsConnection) connectClient() error {
	log.Debug("Connecting to stats..")

	if err := c.statsClient.Connect(); err != nil {
		return err
	}

	log.Debugf("Connected to stats.")

	// store connected state
	atomic.StoreUint32(&c.connected, 1)

	return nil
}

// Disconnect disconnects from Stats API and releases all connection-related resources.
func (c *StatsConnection) Disconnect() {
	if c == nil {
		return
	}

	if c.statsClient != nil {
		c.disconnectClient()
	}
}

func (c *StatsConnection) disconnectClient() {
	if atomic.CompareAndSwapUint32(&c.connected, 1, 0) {
		c.statsClient.Disconnect()
	}
}

// GetSystemStats retrieves VPP system stats.
func (c *StatsConnection) GetSystemStats() (*api.SystemStats, error) {
	stats, err := c.statsClient.DumpStats(SystemStatsPrefix)
	if err != nil {
		return nil, err
	}

	sysStats := &api.SystemStats{}

	for _, stat := range stats {
		switch stat.Name {
		case SystemStats_VectorRate:
			sysStats.VectorRate = scalarStatToFloat64(stat.Data)
		case SystemStats_InputRate:
			sysStats.InputRate = scalarStatToFloat64(stat.Data)
		case SystemStats_LastUpdate:
			sysStats.LastUpdate = scalarStatToFloat64(stat.Data)
		case SystemStats_LastStatsClear:
			sysStats.LastStatsClear = scalarStatToFloat64(stat.Data)
		case SystemStats_Heartbeat:
			sysStats.Heartbeat = scalarStatToFloat64(stat.Data)
		}
	}

	return sysStats, nil
}

// GetErrorStats retrieves VPP error stats.
func (c *StatsConnection) GetErrorStats(names ...string) (*api.ErrorStats, error) {
	var patterns []string
	if len(names) > 0 {
		patterns = make([]string, len(names))
		for i, name := range names {
			patterns[i] = CounterStatsPrefix + name
		}
	} else {
		// retrieve all error counters by default
		patterns = []string{CounterStatsPrefix}
	}
	stats, err := c.statsClient.DumpStats(patterns...)
	if err != nil {
		return nil, err
	}

	var errorStats = &api.ErrorStats{}

	for _, stat := range stats {
		statName := strings.TrimPrefix(stat.Name, CounterStatsPrefix)

		/* TODO: deal with stats that contain '/' in node/counter name
		parts := strings.Split(statName, "/")
		var nodeName, counterName string
		switch len(parts) {
		case 2:
			nodeName = parts[0]
			counterName = parts[1]
		case 3:
			nodeName = parts[0] + parts[1]
			counterName = parts[2]
		}*/

		errorStats.Errors = append(errorStats.Errors, api.ErrorCounter{
			CounterName: statName,
			Value:       errorStatToUint64(stat.Data),
		})
	}

	return errorStats, nil
}

// GetNodeStats retrieves VPP per node stats.
func (c *StatsConnection) GetNodeStats() (*api.NodeStats, error) {
	stats, err := c.statsClient.DumpStats(NodeStatsPrefix)
	if err != nil {
		return nil, err
	}

	nodeStats := &api.NodeStats{}
	var setPerNode = func(perNode []uint64, fn func(c *api.NodeCounters, v uint64)) {
		if nodeStats.Nodes == nil {
			nodeStats.Nodes = make([]api.NodeCounters, len(perNode))
			for i := range perNode {
				nodeStats.Nodes[i].NodeIndex = uint32(i)
			}
		}
		for i, v := range perNode {
			nodeCounters := nodeStats.Nodes[i]
			fn(&nodeCounters, v)
			nodeStats.Nodes[i] = nodeCounters
		}
	}

	for _, stat := range stats {
		switch stat.Name {
		case NodeStats_Clocks:
			setPerNode(reduceSimpleCounterStat(stat.Data), func(c *api.NodeCounters, v uint64) {
				c.Clocks = v
			})
		case NodeStats_Vectors:
			setPerNode(reduceSimpleCounterStat(stat.Data), func(c *api.NodeCounters, v uint64) {
				c.Vectors = v
			})
		case NodeStats_Calls:
			setPerNode(reduceSimpleCounterStat(stat.Data), func(c *api.NodeCounters, v uint64) {
				c.Calls = v
			})
		case NodeStats_Suspends:
			setPerNode(reduceSimpleCounterStat(stat.Data), func(c *api.NodeCounters, v uint64) {
				c.Suspends = v
			})
		}
	}

	return nodeStats, nil
}

// GetInterfaceStats retrieves VPP per interface stats.
func (c *StatsConnection) GetInterfaceStats() (*api.InterfaceStats, error) {
	stats, err := c.statsClient.DumpStats(InterfaceStatsPrefix)
	if err != nil {
		return nil, err
	}

	ifStats := &api.InterfaceStats{}
	var setPerIf = func(perIf []uint64, fn func(c *api.InterfaceCounters, v uint64)) {
		if ifStats.Interfaces == nil {
			ifStats.Interfaces = make([]api.InterfaceCounters, len(perIf))
			for i := range perIf {
				ifStats.Interfaces[i].InterfaceIndex = uint32(i)
			}
		}
		for i, v := range perIf {
			ifCounters := ifStats.Interfaces[i]
			fn(&ifCounters, v)
			ifStats.Interfaces[i] = ifCounters
		}
	}

	for _, stat := range stats {
		switch stat.Name {
		case InterfaceStats_Drops:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.Drops = v
			})
		case InterfaceStats_Punt:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.Punts = v
			})
		case InterfaceStats_IP4:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.IP4 = v
			})
		case InterfaceStats_IP6:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.IP6 = v
			})
		case InterfaceStats_RxNoBuf:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.RxNoBuf = v
			})
		case InterfaceStats_RxMiss:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.RxMiss = v
			})
		case InterfaceStats_RxError:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.RxErrors = v
			})
		case InterfaceStats_TxError:
			setPerIf(reduceSimpleCounterStat(stat.Data), func(c *api.InterfaceCounters, v uint64) {
				c.TxErrors = v
			})
		case InterfaceStats_Rx:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.RxPackets = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.RxBytes = v
			})
		case InterfaceStats_RxUnicast:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.RxUnicast[0] = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.RxUnicast[1] = v
			})
		case InterfaceStats_RxMulticast:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.RxMulticast[0] = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.RxMulticast[1] = v
			})
		case InterfaceStats_RxBroadcast:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.RxBroadcast[0] = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.RxBroadcast[1] = v
			})
		case InterfaceStats_Tx:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.TxPackets = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.TxBytes = v
			})
		case InterfaceStats_TxUnicastMiss:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.TxUnicastMiss[0] = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.TxUnicastMiss[1] = v
			})
		case InterfaceStats_TxMulticast:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.TxMulticast[0] = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.TxMulticast[1] = v
			})
		case InterfaceStats_TxBroadcast:
			per := reduceCombinedCounterStat(stat.Data)
			setPerIf(per[0], func(c *api.InterfaceCounters, v uint64) {
				c.TxBroadcast[0] = v
			})
			setPerIf(per[1], func(c *api.InterfaceCounters, v uint64) {
				c.TxBroadcast[1] = v
			})
		}
	}

	return ifStats, nil
}

func scalarStatToFloat64(stat adapter.Stat) float64 {
	if s, ok := stat.(adapter.ScalarStat); ok {
		return float64(s)
	}
	return 0
}

func errorStatToUint64(stat adapter.Stat) uint64 {
	if s, ok := stat.(adapter.ErrorStat); ok {
		return uint64(s)
	}
	return 0
}

func reduceSimpleCounterStat(stat adapter.Stat) []uint64 {
	if s, ok := stat.(adapter.SimpleCounterStat); ok {
		if len(s) == 0 {
			return []uint64{}
		}
		var per = make([]uint64, len(s[0]))
		for _, w := range s {
			for i, n := range w {
				per[i] += uint64(n)
			}
		}
		return per
	}
	return nil
}

func reduceCombinedCounterStat(stat adapter.Stat) [2][]uint64 {
	if s, ok := stat.(adapter.CombinedCounterStat); ok {
		if len(s) == 0 {
			return [2][]uint64{{}, {}}
		}
		var perPackets = make([]uint64, len(s[0]))
		var perBytes = make([]uint64, len(s[0]))
		for _, w := range s {
			for i, n := range w {
				perPackets[i] += uint64(n.Packets)
				perBytes[i] += uint64(n.Bytes)
			}
		}
		return [2][]uint64{perPackets, perBytes}
	}
	return [2][]uint64{}
}
