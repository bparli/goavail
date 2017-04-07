package main

import (
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/dns"
)

//GoavailConfig struct to capture config file parameters
type GoavailConfig struct {
	Port       int           `toml:"port"`
	Threshold  int           `toml:"failure_threshold"`
	SlackAddr  string        `toml:"slack_addr"`
	Interval   time.Duration `toml:"interval"`
	Members    MemberList
	Route53    dns.Route53
	Cloudflare dns.CloudFlare
}

//MemberList to capture the memberlist/cluster configurations
type MemberList struct {
	MembersPort   int      `toml:"members_port"`
	MinPeersAgree int      `toml:"min_peers_agree"`
	Peers         []string `toml:"peers"`
	LocalAddr     string   `toml:"local_addr"`
	CryptoKey     string   `toml:"crypto_key"`
}

func parseConfig(path string) *GoavailConfig {
	var config GoavailConfig

	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s", err.Error())
	}
	return &config
}
