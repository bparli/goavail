package ipState

import (
	log "github.com/Sirupsen/logrus"
)

func handleTransition(ipAddress string, live bool) {
	if Gm.Clustered && Gm.Members.NumMembers() > 1 {
		err := NotifyIPState(ipAddress, live, false)
		if err != nil {
			log.Errorln("Error Updating state: ", ipAddress, err)
		}
	} else {
		if Gm.MinAgreement > 0 {
			log.Debugln("Running in single mode, BUT need agreement from peers in cluster mode")
		} else {
			log.Debugln("Running in single mode, updating DNS")
			if live == true {
				err := Master.DNS.AddIP(ipAddress, Gm.DryRun)
				if err != nil {
					log.Errorln("Error Adding IP: ", ipAddress, err)
				}
			} else {
				err := Master.DNS.RemoveIP(ipAddress, Gm.DryRun)
				if err != nil {
					log.Errorln("Error Removing IP: ", ipAddress, err)
				}
			}
		}
	}
}
