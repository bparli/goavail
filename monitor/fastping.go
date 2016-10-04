package monitor

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/dns"
	"github.com/tatsushid/go-fastping"
)

type Pingmon struct {
	results      map[string]*response
	p            *fastping.Pinger
	AddressState map[string]int
	dns          dns.DnsProvider
}

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

var master Pingmon

func StartPingMon(dnsConfig *dns.CFlare, threshold int) {
	master.p = fastping.NewPinger()

	master.results = make(map[string]*response)
	master.AddressState = make(map[string]int)

	onRecv, onIdle := make(chan *response), make(chan bool)
	master.p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{addr: addr, rtt: t}
	}
	master.p.OnIdle = func() {
		onIdle <- true
	}

	master.results = make(map[string]*response)
	for _, ip := range dnsConfig.Addresses {
		master.results[ip] = nil
		master.p.AddIP(ip)
		master.AddressState[ip] = 0
	}

	master.p.MaxRTT = time.Second
	master.p.RunLoop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

loop:
	for {
		select {
		case <-c:
			log.Debugln("get interrupted")
			break loop
		case res := <-onRecv:
			if _, ok := master.results[res.addr.String()]; ok {
				master.results[res.addr.String()] = res
				if master.AddressState[res.addr.String()] == 0 {
					continue
				} else if master.AddressState[res.addr.String()] == 1 {
					err := master.dns.AddIP(res.addr.String())
					if err != nil {
						log.Errorln("Error adding IP: ", res.addr.String(), err)
					}
					master.AddressState[res.addr.String()] = 0
				} else {
					master.AddressState[res.addr.String()] -= 1
				}
			}
		case <-onIdle:
			for ipAddr, r := range master.results {
				if r == nil {
					log.Debugln(ipAddr, ": unreachable, ", time.Now())
				} else {
					log.Debugln(ipAddr, ": ", r.rtt, " ", time.Now())
				}
				if master.AddressState[ipAddr] == threshold {
					master.AddressState[ipAddr] += 1
					err := master.dns.RemoveIP(ipAddr)
					if err != nil {
						log.Errorln("Error removing IP: ", ipAddr, err)
					}
				} else if master.AddressState[ipAddr] == threshold+1 {
					continue
				} else {
					master.AddressState[ipAddr] += 1
					master.results[ipAddr] = nil
				}
			}
		case <-master.p.Done():
			if err := master.p.Err(); err != nil {
				log.Debugln("Ping failed:", err)
			}
			break loop
		}
	}
	signal.Stop(c)
	master.p.Stop()
}
