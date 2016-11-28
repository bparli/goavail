package ipState

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/encrypt"
	"github.com/nitro/memberlist"
)

type peersIpStatus struct {
	IpAddr  string
	IpsView map[string]bool
}

type GlobalMap struct {
	IpLive       map[string]bool           // true if an IP is still alive and false otherwise
	peersIpView  map[string]*peersIpStatus //map of monitored IP addresses to each peers' views of them
	MinAgreement int
	Mutex        *sync.RWMutex
	Peers        []string
	LocalAddr    string
	Clustered    bool
	Members      *memberlist.Memberlist
	DryRun       bool
	CryptoKey    string
}

type HttpUpdate struct {
	Peer      string
	IpAddress string
	Live      bool
}

var Gm *GlobalMap

//Initialize the Global Map struct
func InitGM(ipAddresses []string, dryRun bool) {
	m := make(map[string]bool)
	for _, ip := range ipAddresses {
		m[ip] = true
	}

	Gm = &GlobalMap{IpLive: m, Mutex: &sync.RWMutex{}, DryRun: dryRun}
}

//Initialize the state of Peers' IP views to be true
func InitPeersIpViews() {
	Gm.peersIpView = make(map[string]*peersIpStatus)
	for ipAddr, _ := range Gm.IpLive {
		Gm.peersIpView[ipAddr] = &peersIpStatus{IpAddr: ipAddr, IpsView: make(map[string]bool)}
		for _, peer := range Gm.Peers {
			Gm.peersIpView[ipAddr].IpsView[peer] = true
		}
	}
}

//notifyPeers will update all peers on an IP address state change
func notifyPeers(ipAddress string, live bool) error {
	LocalAddr := Gm.LocalAddr
	if Gm.CryptoKey != "" {
		LocalAddr = encrypt.Encrypt([]byte(Gm.CryptoKey), Gm.LocalAddr)
		ipAddress = encrypt.Encrypt([]byte(Gm.CryptoKey), ipAddress)
	}

	for _, peer := range Gm.Peers {
		log.Debugln("Send update to peer:", peer)
		upd := &HttpUpdate{Peer: LocalAddr, IpAddress: ipAddress, Live: live}
		data, err := json.Marshal(upd)
		if err != nil {
			log.Errorln("Error Marshaling", err)
			return err
		}
		buff := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", "http://"+peer, buff)
		if err != nil {
			log.Errorln("Error forming new request", err)
			return err
		}
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

//Update ipstate in Peers' views based on http update
func UpdateGlobalState(ipAddress string, live bool, peer string) (liveCheck int) {
	Gm.Mutex.Lock()
	Gm.peersIpView[ipAddress].IpsView[peer] = live
	liveCheck = checkGlobalState(ipAddress)
	Gm.Mutex.Unlock()
	return
}

func checkGlobalState(ipAddress string) (liveCheck int) {
	liveCheck = 0
	check := Gm.peersIpView[ipAddress]
	for _, v := range check.IpsView {
		if v {
			liveCheck += 1
		} else {
			liveCheck -= 1
		}
	}
	return
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
