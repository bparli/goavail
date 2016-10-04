package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/dns"
	"github.com/bparli/goavail/monitor"
)

func configureCloudflare(config *Config) (*dns.CFlare, error) {
	dnsConfig := dns.CFlare{
		config.DnsDomain,
		config.Proxied,
		config.Addresses,
		config.HostNames}

	return &dnsConfig, nil
}

func main() {
	log.SetLevel(log.DebugLevel)
	opts := parseCommandLine()
	config := parseConfig(*opts.ConfigFile)

	dnsConfig, err := configureCloudflare(&config)
	if err != nil {
		log.Fatalln("Error initializing Cloudflare: ", err)

	}

	log.Debugln(config.Addresses)
	monitor.StartPingMon(dnsConfig, config.Threshold)
}
