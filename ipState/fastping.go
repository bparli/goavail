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

type Pingmon struct {
	Results          map[string]*response
	P                *fastping.Pinger
	AddressFails     map[string]int
	AddressSuccesses map[string]int
	Dns              dns.DnsProvider
	Mutex            *sync.RWMutex
}

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

var Master Pingmon

func StartPingMon(dnsConfig *dns.CFlare, threshold int) {

	Master.P = fastping.NewPinger()
	Master.Dns = dnsConfig
	Master.Mutex = &sync.RWMutex{}

	Master.Results = make(map[string]*response)
	Master.AddressFails = make(map[string]int)
	Master.AddressSuccesses = make(map[string]int)

	onRecv, onIdle := make(chan *response), make(chan bool)
	Master.P.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{addr: addr, rtt: t}
	}
	Master.P.OnIdle = func() {
		onIdle <- true
	}

	Master.Results = make(map[string]*response)
	for _, ip := range dnsConfig.Addresses {
		Master.Results[ip] = nil
		Master.P.AddIP(ip)
		Master.AddressFails[ip] = 0
		Master.AddressSuccesses[ip] = 4 //initialize IPs such that they are already in service at start time
	}

	Master.P.MaxRTT = 2 * time.Second

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
			if Master.AddressSuccesses[res.addr.String()] == 3 {
				log.Infoln("IP Address ", res.addr.String(), " back in service")
				handleTansition(res.addr.String(), true)
			}
			Master.AddressSuccesses[res.addr.String()] += 1
			Master.AddressFails[res.addr.String()] = 0
		case <-onIdle:
			for ipAddr, r := range Master.Results {
				if r == nil {
					log.Debugln(ipAddr, ": unreachable, ", time.Now())
					if Master.AddressFails[ipAddr] == 3 {
						handleTansition(ipAddr, false)
					}
					Master.AddressFails[ipAddr] += 1
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
