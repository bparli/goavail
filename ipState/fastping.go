package ipState

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/dns"
	"github.com/tatsushid/go-fastping"
)

//Pingmon struct to manage monitoring agent context
type Pingmon struct {
	Results          map[string]*response
	P                *fastping.Pinger
	AddressFails     map[string]int
	AddressSuccesses map[string]int
	DNS              dns.Provider
	Mutex            *sync.RWMutex
}

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

//Master pinger
var Master Pingmon

func initMaster(dnsConfig *dns.CFlare, threshold int) {
	Master.P = fastping.NewPinger()
	Master.DNS = dnsConfig
	Master.Mutex = &sync.RWMutex{}
	Master.Results = make(map[string]*response)
	Master.AddressFails = make(map[string]int)
	Master.AddressSuccesses = make(map[string]int)

	Master.Results = make(map[string]*response)
	for _, ip := range dnsConfig.Addresses {
		Master.Results[ip] = nil
		Master.P.AddIP(ip)
		Master.AddressFails[ip] = 0
		Master.AddressSuccesses[ip] = threshold + 1 //initialize IPs such that they are already in service at start time
	}

	Master.P.MaxRTT = 2 * time.Second
}

//StartPingMon to initialize and start ping monitor
func StartPingMon(dnsConfig *dns.CFlare, threshold int) {
	initMaster(dnsConfig, threshold)
	onRecv, onIdle := make(chan *response), make(chan bool)
	Master.P.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{addr: addr, rtt: t}
	}
	Master.P.OnIdle = func() {
		onIdle <- true
	}

	Master.P.RunLoop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

loop:
	for {
		Master.Mutex.RLock()
		select {
		case <-c:
			log.Debugln("get interrupted")
			break loop
		case res := <-onRecv:
			Master.Results[res.addr.String()] = res
			if Master.AddressSuccesses[res.addr.String()] == threshold {
				log.Infoln("IP Address ", res.addr.String(), " back in service")
				handleTransition(res.addr.String(), true)
			}
			Master.AddressSuccesses[res.addr.String()]++
			Master.AddressFails[res.addr.String()] = 0
		case <-onIdle:
			for ipAddr, r := range Master.Results {
				if r == nil {
					log.Debugln(ipAddr, ": unreachable, ", time.Now())
					if Master.AddressFails[ipAddr] == threshold {
						handleTransition(ipAddr, false)
					}
					Master.AddressFails[ipAddr]++
					Master.AddressSuccesses[ipAddr] = 0
				} else {
					log.Debugln(ipAddr, ": ", r.rtt, " ", time.Now())
				}
				Master.Results[ipAddr] = nil
			}
		}
		Master.Mutex.RUnlock()
	}
	signal.Stop(c)
	Master.P.Stop()
}
