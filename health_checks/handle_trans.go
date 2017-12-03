package checks

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ADD    int = 1
	REMOVE int = 2
)

//HandleTransition verifies neccessary peer agreement and takeks appropriate action
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
				//err := Master.DNS.AddIP(ipAddress, Gm.DryRun)
				err := ChangeIP(ipAddress, Gm.DryRun, ADD)
				if err != nil {
					log.Errorln("Error Adding IP: ", ipAddress, err)
				}
			} else {
				//err := Master.DNS.RemoveIP(ipAddress, Gm.DryRun)
				err := ChangeIP(ipAddress, Gm.DryRun, REMOVE)
				if err != nil {
					log.Errorln("Error Removing IP: ", ipAddress, err)
				}
			}
		}
	}
}

func ChangeIP(ipAddress string, dryRun bool, op int) error {
	var err error

	if Gm.Clustered {
		if err = Gm.Dmutex.Lock(); err != nil {
			return errors.New("Error acquiring distributed lock: " + err.Error())
		} else {
			log.Debugln("Acquired distributed lock: ", time.Now())
		}
	}

	if op == ADD {
		err = Master.DNS.AddIP(ipAddress, dryRun)
	} else if op == REMOVE {
		err = Master.DNS.RemoveIP(ipAddress, dryRun)
	}

	if Gm.Clustered {
		// if we get to this point don't return an error unless the actual update call above
		// throws an error. But still log it
		if errUnlock := Gm.Dmutex.UnLock(); errUnlock != nil {
			log.Errorln(errUnlock)
		} else {
			log.Debugln("Released distributed lock: ", time.Now())
		}
	}
	return err
}
