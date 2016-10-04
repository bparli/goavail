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
	results          map[string]*response
	p                *fastping.Pinger
	AddressFails     map[string]int
	AddressSuccesses map[string]int
	dns              dns.DnsProvider
}

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

var master Pingmon

func StartPingMon(dnsConfig *dns.CFlare, threshold int) {
	master.p = fastping.NewPinger()
	master.dns = dnsConfig

	master.results = make(map[string]*response)
	master.AddressFails = make(map[string]int)
	master.AddressSuccesses = make(map[string]int)

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
		master.AddressFails[ip] = 0
		master.AddressSuccesses[ip] = 0
	}

	master.p.MaxRTT = 2 * time.Second
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
			master.results[res.addr.String()] = res
			if master.AddressSuccesses[res.addr.String()] >= 3 {
				log.Infoln("IP Address ", res.addr.String(), " back in service")
				err := master.dns.AddIP(res.addr.String())
				if err != nil {
					log.Errorln("Error adding IP: ", res.addr.String(), err)
				}
			}
			master.AddressSuccesses[res.addr.String()] += 1
			master.AddressFails[res.addr.String()] = 0
		case <-onIdle:
			for ipAddr, r := range master.results {
				if r == nil {
					log.Debugln(ipAddr, ": unreachable, ", time.Now())
					if master.AddressFails[ipAddr] >= 3 {
						err := master.dns.RemoveIP(ipAddr)
						if err != nil {
							log.Errorln("Error removing IP: ", ipAddr, err)
						}
					}
					master.AddressFails[ipAddr] += 1
					master.AddressSuccesses[ipAddr] = 0
				} else {
					log.Debugln(ipAddr, ": ", r.rtt, " ", time.Now())
				}
				master.results[ipAddr] = nil
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
