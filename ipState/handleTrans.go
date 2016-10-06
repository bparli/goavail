package ipState

import log "github.com/Sirupsen/logrus"

func handleTansition(ipAddress string, live bool) {
	if Gm.Clustered && Gm.Members.NumMembers() > 1 {
		err := NotifyIpState(ipAddress, live, false)
		if err != nil {
			log.Errorln("Error Updating state: ", ipAddress, err)
		}
	} else {
		if live == true {
			err := master.dns.AddIP(ipAddress)
			if err != nil {
				log.Errorln("Error Adding IP: ", ipAddress, err)
			}
		} else {
			err := master.dns.RemoveIP(ipAddress)
			if err != nil {
				log.Errorln("Error Removing IP: ", ipAddress, err)
			}
		}

	}
}
