package main

import (
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/dns"
	"github.com/bparli/goavail/httpService"
	"github.com/bparli/goavail/ipState"
)

func configureCloudflare(config *GoavailConfig) (*dns.CFlare, error) {
	dnsConfig := dns.CFlare{
		config.DnsDomain,
		config.Proxied,
		config.Addresses,
		config.HostNames}

	return &dnsConfig, nil
}

func main() {
	runtime.GOMAXPROCS(2)
	log.SetLevel(log.DebugLevel)
	opts := parseCommandLine()
	config := parseConfig(*opts.ConfigFile)

	dnsConfig, err := configureCloudflare(config)
	if err != nil {
		log.Fatalln("Error initializing Cloudflare: ", err)

	}

	ipState.InitGM(config.Addresses)
	if len(config.Peers) > 0 && config.LocalAddr != "" {
		go httpService.UpdatesListener(config.LocalAddr)
		log.Debugln("Running in Cluster mode")
		ipState.Gm.Clustered = true
		ipState.Gm.Peers = config.Peers
		ipState.Gm.LocalAddr = config.LocalAddr
		ipState.InitMembersList(config.LocalAddr, config.Peers)
	} else {
		log.Debugln("Running in Single Node mode.  Need local_addr and peers to be set to run in Cluster Mode")
		ipState.Gm.Clustered = false
	}

	log.Debugln(config)
	ipState.StartPingMon(dnsConfig, config.Threshold)
}
