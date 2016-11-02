package main

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/ipState"
	"github.com/nitro/memberlist"
)

func initMembersList(localAddr string, peer []string, membersPort int) {
	memberlistConfig := memberlist.DefaultWANConfig()
	localIP := strings.Split(localAddr, ":")[0]
	memberlistConfig.AdvertiseAddr = localIP
	memberlistConfig.AdvertisePort = membersPort
	memberlistConfig.BindPort = membersPort

	var err error
	ipState.Gm.Members, err = memberlist.Create(memberlistConfig)
	if err != nil {
		log.Errorln("Failed to create memberlist: " + err.Error())
	}

	// Join an existing cluster by specifying at least one known member.
	var memberIPs []string
	for _, peer := range ipState.Gm.Peers {
		peerIP := strings.Split(peer, ":")[0]
		memberIPs = append(memberIPs, peerIP)
	}
	_, err = ipState.Gm.Members.Join(memberIPs)
	if err != nil {
		log.Errorln("Failed to join cluster: " + err.Error())
	}
}
