package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//GoavailOpts type to pass input parameters
type GoavailOpts struct {
	Command    *string
	ConfigFile *string
	DryRun     *bool
	Debug      *bool
	Type       *string
	DNS        *string
}

func parseCommandLine() *GoavailOpts {
	var opts GoavailOpts

	kingpin.CommandLine.Help = "A monitoring and DNS failover tool."
	opts.ConfigFile = kingpin.Flag("config-file", "The configuration TOML file path").Short('f').Default("goavail.toml").String()
	kingpin.Command("monitor", "Monitor set of Public IP Addresses in goavail.toml")
	kingpin.Flag("laddr", "The port to listen for updates on from peers (Cluster mode only)").Short('l').Default("8081").String()
	opts.DryRun = kingpin.Flag("dry-run", "Is this a dry run?").Short('d').Default("true").Bool()
	opts.Debug = kingpin.Flag("debug", "Set for Debug mode").Short('b').Default("false").Bool()
	opts.Type = kingpin.Flag("type", "Type of monitoring, ip (ping) or tcp based").Short('t').Default("ip").String()
	opts.DNS = kingpin.Flag("dns-provider", "Set DNS Provider, either cloudflare or route53").Short('p').Default("cloudflare").String()

	command := kingpin.Parse()
	opts.Command = &command

	log.Debugln("Using ", opts.ConfigFile, " for configs")

	return &opts
}
