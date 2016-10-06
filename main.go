package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

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

func loadMonitor(configFile string) {
	config := parseConfig(configFile)

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
	go ipState.StartPingMon(dnsConfig, config.Threshold)
}

func reloadMonitor(configFile string) {
	config := parseConfig(configFile)

	dnsConfig, err := configureCloudflare(config)
	if err != nil {
		log.Fatalln("Error initializing Cloudflare: ", err)

	}

	ipState.Gm.Mutex.Lock()
	if len(config.Peers) > 0 && config.LocalAddr != "" && ipState.Gm.Clustered == false { //if we aren't currently clustered but want to be
		go httpService.UpdatesListener(config.LocalAddr)
		log.Debugln("Running in Cluster mode")
		ipState.Gm.Clustered = true
		ipState.Gm.Peers = config.Peers
		ipState.Gm.LocalAddr = config.LocalAddr
		ipState.InitMembersList(config.LocalAddr, config.Peers)
	} else if len(config.Peers) == 0 && config.LocalAddr == "" && ipState.Gm.Clustered == true { //if we are currently clustered but don't want to be
		log.Debugln("Running in Single Node mode.  Need local_addr and peers to be set to run in Cluster Mode")
		ipState.Gm.Clustered = false
		ipState.Gm.Members.Shutdown()
	}
	ipState.Master.Dns = dnsConfig
	ipState.Gm.Mutex.Unlock()

	ipState.Master.Mutex.Lock()
	for _, ip := range dnsConfig.Addresses {
		ipState.Master.Results[ip] = nil
		ipState.Master.P.AddIP(ip)
		ipState.Master.AddressFails[ip] = 0
		ipState.Master.AddressSuccesses[ip] = 4 //initialize IPs such that they are already in service at start time
	}
	ipState.Master.Mutex.Unlock()
	log.Debugln(config)
}

func main() {
	runtime.GOMAXPROCS(2)
	log.SetLevel(log.DebugLevel)
	opts := parseCommandLine()
	if *opts.Command == "monitor" {
		loadMonitor(*opts.ConfigFile)
	}
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP)
	for {
		<-s
		ipState.Master.P.Done()
		reloadMonitor(*opts.ConfigFile)
	}
}
