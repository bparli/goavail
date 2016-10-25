package ipState

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/nitro/memberlist"
)

type GlobalMap struct {
	IpLive    map[string]bool // true if an IP is still alive and false otherwise
	Mutex     *sync.RWMutex
	Gossip    *memberlist.Memberlist
	Peers     []string
	LocalAddr string
	Clustered bool
	Members   *memberlist.Memberlist
	DryRun    bool
}

type HttpUpdate struct {
	Peer      string
	IpAddress string
	Live      bool
}

var Gm *GlobalMap

func InitGM(ipAddresses []string, dryRun bool) {
	m := make(map[string]bool)
	for _, ip := range ipAddresses {
		m[ip] = true
	}

	Gm = &GlobalMap{IpLive: m, Mutex: &sync.RWMutex{}, DryRun: dryRun}
}

//NotifyPeers will update all peers on an IP address state change
func notifyPeers(ipAddress string, live bool) error {
	for _, peer := range Gm.Peers {
		log.Debugln("Send update to peer:", peer)
		upd := &HttpUpdate{Peer: Gm.LocalAddr, IpAddress: ipAddress, Live: live}
		data, err := json.Marshal(upd)
		if err != nil {
			log.Errorln("Error Marshaling", err)
			return err
		}
		buff := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", "http://"+peer, buff)
		req.Close = true
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Errorln("HTTP Post Error", err)
			return err
		}
		defer resp.Body.Close()
	}
	log.Debugln("Sent update to Peers", Gm.Peers)
	return nil
}

func NotifyIpState(ipAddress string, live bool, peerUpdate bool) error {
	Gm.Mutex.RLock()
	if Gm.IpLive[ipAddress] == live && peerUpdate == true { //if we have seen the failure and received a notification from a peer that they've seen the failure
		Gm.Mutex.RUnlock()
		log.Debugln("Received agreement from Peer.  Updating IP Pool")
		if live == false {
			err := Master.Dns.RemoveIp(ipAddress, Gm.DryRun)
			if err != nil {
				log.Errorln("Error Removing IP: ", ipAddress, err)
			}
		} else {
			err := Master.Dns.AddIp(ipAddress, Gm.DryRun)
			if err != nil {
				log.Errorln("Error Adding IP: ", ipAddress, err)
			}
		}
	} else if peerUpdate == false { //we haven't received a message from a peer yet so just update the state and notify peers
		Gm.Mutex.RUnlock()
		Gm.Mutex.Lock()
		Gm.IpLive[ipAddress] = live
		Gm.Mutex.Unlock()
		log.Debugln("Notifying peers: ", ipAddress, live)
		err := notifyPeers(ipAddress, live)
		if err != nil {
			log.Errorln("Error notifying Peers", err)
		}
	} else {
		Gm.Mutex.RUnlock()
	}

	return nil
}
