package telemetry

import (
	"context"
	"fmt"
	"time"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/vpp"
	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/telemetry/vppcalls"
)

type statsPollerServer struct {
	handler vppcalls.TelemetryVppAPI

	log logging.Logger
}

func (s *statsPollerServer) PollStats(req *configurator.PollStatsRequest, svr configurator.StatsPoller_PollStatsServer) error {
	var pollSeq uint32

	period := time.Duration(req.PeriodSec) * time.Second
	tick := time.NewTicker(period)

	s.log.Debugf("starting to poll stats every %v", period)
	for {
		select {
		case <-tick.C:
			pollSeq++
			s.log.WithField("seq", pollSeq).Debugf("polling stats..")

			vppStatsCh := make(chan vpp.Stats)
			var vppStatsErr error
			go func() {
				vppStatsErr = s.streamVppStats(vppStatsCh)
			}()
			for vppStats := range vppStatsCh {
				VppStats := vppStats
				s.log.Debugf("sending vpp stats: %v", VppStats)

				if err := svr.Send(&configurator.PollStatsResponse{
					PollSeq: pollSeq,
					Stats: &configurator.Stats{
						Stats: &configurator.Stats_VppStats{VppStats: &VppStats},
					},
				}); err != nil {
					s.log.Errorf("sending stats failed: %v", err)
					return nil
				}
			}

			if vppStatsErr != nil {
				s.log.Errorf("polling vpp stats failed: %v", vppStatsErr)
				return vppStatsErr
			}
		}
	}
}

func (s *statsPollerServer) streamVppStats(ch chan vpp.Stats) error {
	ctx := context.Background()
	ifStats, err := s.handler.GetInterfaceStats(ctx)
	if err != nil {
		return err
	} else if ifStats == nil {
		return fmt.Errorf("interface stats not avaiable")
	}

	s.log.Debugf("streaming %d interface stats", len(ifStats.Interfaces))

	for _, iface := range ifStats.Interfaces {
		vppStats := vpp.Stats{
			Interface: &vpp_interfaces.InterfaceStats{
				Name:        iface.InterfaceName,
				Rx:          convertInterfaceCombined(iface.Rx),
				Tx:          convertInterfaceCombined(iface.Tx),
				RxUnicast:   convertInterfaceCombined(iface.RxUnicast),
				RxMulticast: convertInterfaceCombined(iface.RxMulticast),
				RxBroadcast: convertInterfaceCombined(iface.RxBroadcast),
				TxUnicast:   convertInterfaceCombined(iface.TxUnicast),
				TxMulticast: convertInterfaceCombined(iface.TxMulticast),
				TxBroadcast: convertInterfaceCombined(iface.TxBroadcast),
				RxError:     iface.RxErrors,
				TxError:     iface.TxErrors,
				RxNoBuf:     iface.RxNoBuf,
				RxMiss:      iface.RxMiss,
				Drops:       iface.Drops,
				Punts:       iface.Punts,
				Ip4:         iface.IP4,
				Ip6:         iface.IP6,
				Mpls:        iface.Mpls,
			},
		}
		ch <- vppStats
	}

	close(ch)

	return nil
}

func convertInterfaceCombined(c api.InterfaceCounterCombined) *vpp_interfaces.InterfaceStats_CombinedCounter {
	return &vpp_interfaces.InterfaceStats_CombinedCounter{
		Bytes:   c.Bytes,
		Packets: c.Packets,
	}

}
