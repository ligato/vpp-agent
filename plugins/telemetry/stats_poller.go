package telemetry

import (
	"context"
	"fmt"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

type statsPollerServer struct {
	handler vppcalls.TelemetryVppAPI
	ifIndex ifaceidx.IfaceMetadataIndex

	log logging.Logger
}

func (s *statsPollerServer) PollStats(req *configurator.PollStatsRequest, svr configurator.StatsPollerService_PollStatsServer) error {
	var pollSeq uint32

	if req.PeriodSec == 0 {
		return status.Error(codes.InvalidArgument, "PeriodSec must be greater than 0")
	}
	period := time.Duration(req.PeriodSec) * time.Second

	tick := time.NewTicker(period)
	defer tick.Stop()

	s.log.Debugf("starting to poll stats every %v", period)
	for {
		pollSeq++
		s.log.WithField("seq", pollSeq).Debugf("polling stats..")

		vppStatsCh := make(chan vpp.Stats)
		var err error
		go func() {
			err = s.streamVppStats(vppStatsCh)
			close(vppStatsCh)
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
		if err != nil {
			s.log.Errorf("polling vpp stats failed: %v", err)
			return err
		}

		<-tick.C
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
		name, _, exists := s.ifIndex.LookupBySwIfIndex(iface.InterfaceIndex)
		if !exists {
			// fallback to internal name
			name = iface.InterfaceName
		}
		vppStats := vpp.Stats{
			Interface: &vpp_interfaces.InterfaceStats{
				Name:        name,
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
	return nil
}

func convertInterfaceCombined(c govppapi.InterfaceCounterCombined) *vpp_interfaces.InterfaceStats_CombinedCounter {
	return &vpp_interfaces.InterfaceStats_CombinedCounter{
		Bytes:   c.Bytes,
		Packets: c.Packets,
	}
}
