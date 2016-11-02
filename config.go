package main

import (
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

type GoavailConfig struct {
	Addresses     []string `toml:"ip_addresses"`
	Threshold     int      `toml:"failure_threshold"`
	DnsDomain     string   `toml:"dns_domain"`
	HostNames     []string `toml:"hostnames"`
	Proxied       bool     `toml:"dns_proxied"`
	Peers         []string `toml:"peers"`
	LocalAddr     string   `toml:"local_addr"`
	SlackAddr     string   `toml:"slack_addr"`
	MembersPort   int      `toml:"members_port"`
	MinPeersAgree int      `toml:"min_peers_agree"`
}

func parseConfig(path string) *GoavailConfig {
	var config GoavailConfig

	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s", err.Error())
	}
	return &config
}
