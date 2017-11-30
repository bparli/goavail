package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bparli/goavail/dns"
	checks "github.com/bparli/goavail/health_checks"
	"github.com/bparli/goavail/http_service"
	"github.com/bparli/goavail/notify"
	log "github.com/sirupsen/logrus"
)

func configDNS(config *GoavailConfig, dnsProvider string) dns.Provider {
	var dnsConfig dns.Provider
	if dnsProvider == "route53" {
		log.Debugln("Using Route53 DNS provider")
		if config.Route53.DNSDomain == "" {
			log.Fatalln("No AWS Domain set")
		}
		r53 := dns.ConfigureRoute53(config.Route53.DNSDomain, config.Route53.AWSRegion, config.Route53.TTL,
			config.Route53.AWSZoneID, config.Route53.Addresses, config.Route53.Hostnames)
		dnsConfig = r53

	} else {
		log.Debugln("Using Cloudflare DNS provider")
		if config.Cloudflare.DNSDomain == "" {
			log.Fatalln("No Cloudflare Domain set")
		}
		cf := dns.ConfigureCloudflare(config.Cloudflare.DNSDomain, config.Cloudflare.Proxied, config.Cloudflare.Addresses, config.Cloudflare.Hostnames)
		dnsConfig = cf
	}
	return dnsConfig
}

func loadMonitor(opts *GoavailOpts) {
	config := parseConfig(*opts.ConfigFile)
	if config.SlackAddr != "" {
		notify.InitSlack(config.SlackAddr)
	}

	dnsConfig := configDNS(config, *opts.DNS)
	addrs := dnsConfig.GetAddrs()

	checks.InitGM(addrs, *opts.DryRun)
	if len(config.Members.Peers) > 0 && config.Members.LocalAddr != "" {
		go httpService.UpdatesListener(config.Members.LocalAddr)
		log.Debugln("Running in Cluster mode")
		checks.Gm.Clustered = true
		checks.Gm.Peers = config.Members.Peers
		checks.Gm.LocalAddr = config.Members.LocalAddr
		checks.Gm.MinAgreement = config.Members.MinPeersAgree
		checks.InitPeersIPViews()
		initMembersList(config.Members.LocalAddr, config.Members.Peers, config.Members.MembersPort)
	} else {
		log.Debugln("Running in Single Node mode.  Need local_addr and peers to be set to run in Cluster Mode")
		checks.Gm.Clustered = false
	}

	if config.Members.CryptoKey != "" {
		checks.Gm.CryptoKey = config.Members.CryptoKey
	}
	checks.Gm.Type = *opts.Type

	checks.NewChecks(dnsConfig, config.Threshold, config.Interval, config.Port)
	if *opts.Type == "ip" {
		log.Debugln("Running IP Ping monitor")
		go checks.StartPingMon(config.Threshold)
	} else if *opts.Type == "tcp" {
		log.Debugln("Running TCP Health Checks monitor")
		go checks.StartTCPChecks(config.Threshold)
	}

}

func reloadMonitor(opts *GoavailOpts) {
	config := parseConfig(*opts.ConfigFile)
	dnsConfig := configDNS(config, *opts.DNS)

	checks.Gm.Mutex.Lock()
	defer checks.Gm.Mutex.Unlock()

	checks.Gm.MinAgreement = config.Members.MinPeersAgree

	checks.Master.DNS = dnsConfig

	for _, ip := range dnsConfig.GetAddrs() {
		checks.Master.Results[ip] = nil
		checks.Master.P.AddIP(ip)
		checks.Master.AddressFails[ip] = 0
		checks.Master.AddressSuccesses[ip] = config.Threshold + 1 //initialize IPs such that they are already in service at start time
	}
	log.Debugln(config)
}

func main() {
	//runtime.GOMAXPROCS(2)
	opts := parseCommandLine()
	if *opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if *opts.Command == "monitor" {
		loadMonitor(opts)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP)
	for {
		<-s
		checks.Master.P.Done()
		reloadMonitor(opts)
	}
}
