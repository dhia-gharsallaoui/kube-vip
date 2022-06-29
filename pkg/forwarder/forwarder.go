package forwarder

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Forwarder interface {
	Start(context.Context) error
}

type forwarder struct {
	srcAddress           string
	srcPort              int
	dstAddress           string
	dstPort              int
	ipvs                 bool
	ipvsForwardingMethod string
}

func NewForwarder(srcAddress string, srcPort int, dstAddress string, dstPort int, ipvs bool, ipvsForwardingMethod string) Forwarder {
	return &forwarder{
		srcAddress:           srcAddress,
		srcPort:              srcPort,
		dstAddress:           dstAddress,
		dstPort:              dstPort,
		ipvs:                 ipvs,
		ipvsForwardingMethod: ipvsForwardingMethod,
	}
}

func (f *forwarder) Start(ctx context.Context) error {
	log.Infoln("starting forwarder nftables")
	// If backend address is quad-zero route, early return
	// (destination address is unknown, can't setup forwarding).
	// If at some point we want to prepare forwarding none the less we would have to pick up any interface
	// address and continue with it. If the user want to profit from different vip port and backend port
	// it's probably easier the user define the backend address instead
	if f.dstAddress == "0.0.0.0" {
		log.Infof("dest address is 0.0.0.0")
		fmt.Println(f.dstAddress)
		return nil
	}

	log.Infoln("here")
	go func(ctx context.Context, stop func() error) {
		<-ctx.Done()
		if err := stop(); err != nil {
			log.Errorf("forwarder stop error: %v", err)
		}
	}(ctx, f.Stop)
	if f.ipvs {
		log.Infoln("starting IPVS")
		return startIPVS(f.srcAddress, f.srcPort, f.dstAddress, f.dstPort, f.ipvsForwardingMethod)
	}

	log.Infoln("starting NFTables")
	return startNFTables(ctx, f.srcAddress, f.srcPort, f.dstAddress, f.dstPort)

}

func (f *forwarder) Stop() error {
	log.Infoln("stopping forwarder")
	if f.dstAddress == "0.0.0.0" {
		return nil
	}
	if f.ipvs {

		log.Infoln("stopping IPVS")
		return stopIPVS(f.srcAddress, f.srcPort, f.dstAddress, f.dstPort, f.ipvsForwardingMethod)
	}

	log.Infoln("stopping NFTables")
	return stopNFTables(f.srcAddress, f.srcPort, f.dstAddress, f.dstPort)
}
