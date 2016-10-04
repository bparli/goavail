package main

import (
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

type GoavailOpts struct {
	Command    *string
	ConfigFile *string
	IP         *string
}

func parseCommandLine() *GoavailOpts {
	var opts GoavailOpts

	kingpin.CommandLine.Help = "An monitoring and DNS failover tool."
	opts.ConfigFile = kingpin.Flag("config-file", "The configuration TOML file path").Short('f').Default("goavail.toml").String()
	kingpin.Command("monitor", "Monitor set of Public IP Addresses in goavail.toml")

	command := kingpin.Parse()
	opts.Command = &command

	log.Debugln("Using ", opts.ConfigFile, " for configs")

	return &opts
}
