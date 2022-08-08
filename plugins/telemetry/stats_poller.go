package telemetry

import (
	"context"
	"fmt"
	"time"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

type statsPollerServer struct {
	configurator.UnimplementedStatsPollerServiceServer

	handler vppcalls.TelemetryVppAPI
	ifIndex ifaceidx.IfaceMetadataIndex

	log logging.Logger
}

func (s *statsPollerServer) PollStats(req *configurator.PollStatsRequest, svr configurator.StatsPollerService_PollStatsServer) error {
	if req.GetPeriodSec() == 0 && req.GetNumPolls() > 1 {
		return status.Error(codes.InvalidArgument, "period must be > 0 if number of polls is > 1")
	}
	if s.handler == nil {
		return status.Errorf(codes.Unavailable, "VPP telemetry handler not available")
	}

	ctx := svr.Context()

	streamStats := func(pollSeq uint32) (err error) {
		vppStatsCh := make(chan *vpp.Stats)
		go func() {
			err = s.streamVppStats(ctx, vppStatsCh)
			close(vppStatsCh)
		}()
		for vppStats := range vppStatsCh {
			VppStats := proto.Clone(vppStats).(*vpp.Stats)
			s.log.Debugf("sending vpp stats: %v", VppStats)

			if err := svr.Send(&configurator.PollStatsResponse{
				PollSeq: pollSeq,
				Stats: &configurator.Stats{
					Stats: &configurator.Stats_VppStats{VppStats: VppStats},
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
		return nil
	}

	if req.GetPeriodSec() == 0 {
		return streamStats(0)
	}

	period := time.Duration(req.GetPeriodSec()) * time.Second
	s.log.Debugf("start polling stats every %v", period)

	tick := time.NewTicker(period)
	defer tick.Stop()

	for pollSeq := uint32(1); ; pollSeq++ {
		s.log.WithField("seq", pollSeq).Debugf("polling stats..")

		if err := streamStats(pollSeq); err != nil {
			return err
		}

		if req.GetNumPolls() > 0 && pollSeq >= req.GetNumPolls() {
			s.log.Debugf("reached %d pollings", req.GetNumPolls())
			return nil
		}

		select {
		case <-tick.C:
			// period passed
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *statsPollerServer) streamVppStats(ctx context.Context, ch chan *vpp.Stats) error {
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
		vppStats := &vpp.Stats{
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

		select {
		case ch <- vppStats:
			// stats sent
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func convertInterfaceCombined(c govppapi.InterfaceCounterCombined) *vpp_interfaces.InterfaceStats_CombinedCounter {
	return &vpp_interfaces.InterfaceStats_CombinedCounter{
		Bytes:   c.Bytes,
		Packets: c.Packets,
	}
}
