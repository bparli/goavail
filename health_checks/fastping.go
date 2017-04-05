package checks

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

type response struct {
	netAddr *net.IPAddr
	rtt     time.Duration
	tcpAddr string
}

//StartPingMon to initialize and start ping monitor
func StartPingMon(threshold int) {
	Master.P.MaxRTT = Master.Interval
	onRecv, onIdle := make(chan *response), make(chan bool)
	Master.P.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{netAddr: addr, rtt: t}
	}
	Master.P.OnIdle = func() {
		onIdle <- true
	}

	Master.P.RunLoop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	log.Debugln("Starting pingmon")

loop:
	for {
		Master.Mutex.RLock()
		select {
		case <-c:
			log.Debugln("get interrupted")
			break loop
		case res := <-onRecv:
			Master.Results[res.netAddr.String()] = res
			if Master.AddressSuccesses[res.netAddr.String()] == threshold {
				log.Infoln("IP Address ", res.netAddr.String(), " back in service")
				handleTransition(res.netAddr.String(), true)
			}
			Master.AddressSuccesses[res.netAddr.String()]++
			Master.AddressFails[res.netAddr.String()] = 0
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
