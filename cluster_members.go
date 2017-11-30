package main

import (
	"strings"

	checks "github.com/bparli/goavail/health_checks"
	"github.com/hashicorp/memberlist"
	log "github.com/sirupsen/logrus"
)

func initMembersList(localAddr string, peer []string, membersPort int) {
	memberlistConfig := memberlist.DefaultWANConfig()
	localIP := strings.Split(localAddr, ":")[0]
	memberlistConfig.AdvertiseAddr = localIP
	memberlistConfig.AdvertisePort = membersPort
	memberlistConfig.BindPort = membersPort

	var err error
	checks.Gm.Members, err = memberlist.Create(memberlistConfig)
	if err != nil {
		log.Errorln("Failed to create memberlist: " + err.Error())
	}

	// Join an existing cluster by specifying at least one known member.
	var memberIPs []string
	for _, peer := range checks.Gm.Peers {
		peerIP := strings.Split(peer, ":")[0]
		memberIPs = append(memberIPs, peerIP)
	}
	_, err = checks.Gm.Members.Join(memberIPs)
	if err != nil {
		log.Errorln("Failed to join cluster: " + err.Error())
	}
}
