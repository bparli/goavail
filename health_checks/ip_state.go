package checks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/bparli/goavail/encrypt"
	"github.com/hashicorp/memberlist"
	log "github.com/sirupsen/logrus"
)

type peersIPStatus struct {
	IPAddr  string
	IpsView map[string]bool
}

//GlobalMap struct for managing network of monitoring agents and monitored IPs
type GlobalMap struct {
	IPLive       map[string]bool           // true if an IP is still alive and false otherwise
	peersIPView  map[string]*peersIPStatus //map of monitored IP addresses to each peers' views of them
	MinAgreement int
	Mutex        *sync.RWMutex
	Peers        []string
	LocalAddr    string
	Clustered    bool
	Members      *memberlist.Memberlist
	DryRun       bool
	CryptoKey    string
	Type         string
}

//HTTPUpdate to manage peer monitoring agent updates
type HTTPUpdate struct {
	Peer      string
	IPAddress string
	Live      bool
}

//Gm var for the global map of monitored IPs
var Gm *GlobalMap

//InitGM - initialize the Global Map struct
func InitGM(ipAddresses []string, dryRun bool) {
	m := make(map[string]bool)
	for _, ip := range ipAddresses {
		m[ip] = true
	}

	Gm = &GlobalMap{IPLive: m, Mutex: &sync.RWMutex{}, DryRun: dryRun}
}

//InitPeersIPViews - initialize the state of Peers' IP views to be true
func InitPeersIPViews() {
	Gm.peersIPView = make(map[string]*peersIPStatus)
	for ipAddr := range Gm.IPLive {
		Gm.peersIPView[ipAddr] = &peersIPStatus{IPAddr: ipAddr, IpsView: make(map[string]bool)}
		for _, peer := range Gm.Peers {
			Gm.peersIPView[ipAddr].IpsView[peer] = true
		}
	}
}

//notifyPeers will update all peers on an IP address state change
func notifyPeers(ipAddress string, live bool) error {
	LocalAddr := Gm.LocalAddr
	if Gm.CryptoKey != "" {
		LocalAddr = encrypt.Encrypt([]byte(Gm.CryptoKey), Gm.LocalAddr)
		ipAddress = encrypt.Encrypt([]byte(Gm.CryptoKey), ipAddress)
		log.Debugln("Payload Encrypted")
	}

	for _, peer := range Gm.Peers {
		log.Debugln("Send update to peer:", peer)
		upd := &HTTPUpdate{Peer: LocalAddr, IPAddress: ipAddress, Live: live}
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

//UpdateGlobalState - update ipstate in Peers' views based on http update
func UpdateGlobalState(ipAddress string, live bool, peer string) (liveCheck int) {
	Gm.Mutex.Lock()
	Gm.peersIPView[ipAddress].IpsView[peer] = live
	liveCheck = checkGlobalState(ipAddress)
	Gm.Mutex.Unlock()
	return
}

func checkGlobalState(ipAddress string) (liveCheck int) {
	liveCheck = 0
	check := Gm.peersIPView[ipAddress]
	for _, v := range check.IpsView {
		if v {
			liveCheck++
		} else {
			liveCheck--
		}
	}
	return
}

//NotifyIPState - notify peer agents of change in state of monitored IP
func NotifyIPState(ipAddress string, live bool, peerUpdate bool) error {
	Gm.Mutex.RLock()
	if Gm.IPLive[ipAddress] == live && peerUpdate == true { //if we have seen the failure and received a notification from a peer that they've seen the failure
		Gm.Mutex.RUnlock()
		log.Debugln("Received agreement from Peer.  Updating IP Pool")
		if live == false {
			err := Master.DNS.RemoveIP(ipAddress, Gm.DryRun)
			if err != nil {
				log.Errorln("Error Removing IP: ", ipAddress, err)
			}
		} else {
			err := Master.DNS.AddIP(ipAddress, Gm.DryRun)
			if err != nil {
				log.Errorln("Error Adding IP: ", ipAddress, err)
			}
		}
	} else if peerUpdate == false { //we haven't received a message from a peer yet so just update the state and notify peers
		Gm.Mutex.RUnlock()
		Gm.Mutex.Lock()
		Gm.IPLive[ipAddress] = live
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
