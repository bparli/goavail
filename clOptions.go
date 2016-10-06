package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type GoavailOpts struct {
	Command    *string
	ConfigFile *string
}

func parseCommandLine() *GoavailOpts {
	var opts GoavailOpts

	kingpin.CommandLine.Help = "A monitoring and DNS failover tool."
	opts.ConfigFile = kingpin.Flag("config-file", "The configuration TOML file path").Short('f').Default("goavail.toml").String()
	kingpin.Command("monitor", "Monitor set of Public IP Addresses in goavail.toml")
	kingpin.Flag("laddr", "The port to listen for updates on from peers (Cluster mode only)").Short('l').Default("8081").String()

	command := kingpin.Parse()
	opts.Command = &command

	log.Debugln("Using ", opts.ConfigFile, " for configs")

	return &opts
}
