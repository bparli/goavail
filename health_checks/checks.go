package checks

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bparli/goavail/dns"
	log "github.com/sirupsen/logrus"
	fastping "github.com/tatsushid/go-fastping"
)

//HealthTracker struct to manage monitoring agent context
type HealthTracker struct {
	Results          map[string]*response
	P                *fastping.Pinger
	TCPChecks        map[string]string
	AddressFails     map[string]int
	AddressSuccesses map[string]int
	DNS              dns.Provider
	Mutex            *sync.RWMutex
	Interval         time.Duration
}

//Master pinger
var Master *HealthTracker

//NewChecks initializes the *HealthTracker struct for health checks
func NewChecks(dnsConfig dns.Provider, threshold int, interval time.Duration, port int) {
	Master = &HealthTracker{
		P:                fastping.NewPinger(),
		DNS:              dnsConfig,
		Results:          make(map[string]*response),
		AddressFails:     make(map[string]int),
		AddressSuccesses: make(map[string]int),
		TCPChecks:        make(map[string]string),
		Interval:         interval * time.Second,
		Mutex:            &sync.RWMutex{}}

	for _, ip := range dnsConfig.GetAddrs() {
		Master.Results[ip] = nil
		Master.P.AddIP(ip)
		Master.AddressFails[ip] = 0
		Master.AddressSuccesses[ip] = threshold + 1 //initialize IPs such that they are already in service at start time
		if strings.Compare(Gm.Type, "tcp") == 0 {
			Master.TCPChecks[ip] = ip + ":" + strconv.Itoa(port)
		}
	}
}

//StartTCPChecks to run the registered tcp based health checks in separate goroutines
func StartTCPChecks(threshold int) {
	onSuccess, onFail := make(chan *response), make(chan *response)
	for _, tcpAddr := range Master.TCPChecks {
		go runTCPChecks(tcpAddr, onSuccess, onFail)
	}
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
		case res := <-onSuccess:
			ip := strings.Split(res.tcpAddr, ":")[0]
			Master.Results[ip] = res
			if Master.AddressSuccesses[ip] == threshold {
				log.Infoln("IP Address ", res, " back in service")
				handleTransition(ip, true)
			}
			Master.AddressSuccesses[ip]++
			Master.AddressFails[ip] = 0
			log.Debugln(res.tcpAddr, ": ", time.Now())
		case res := <-onFail:
			ip := strings.Split(res.tcpAddr, ":")[0]
			Master.Results[ip] = nil
			log.Debugln(res.tcpAddr, ": unreachable, ", time.Now())
			if Master.AddressFails[ip] == threshold {
				handleTransition(ip, false)
			}
			Master.AddressFails[ip]++
			Master.AddressSuccesses[ip] = 0
		}
		Master.Mutex.RUnlock()
	}
	signal.Stop(c)
}

func runTCPChecks(tcpAddr string, onSucess chan *response, onFail chan *response) {
	for {
		check := tcpChecker(tcpAddr)
		if check != nil {
			onFail <- &response{tcpAddr: tcpAddr}
		} else {
			onSucess <- &response{tcpAddr: tcpAddr}
		}
		time.Sleep(Master.Interval)
	}
}

func tcpChecker(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return errors.New("connection to " + addr + " failed")
	}
	conn.Close()
	return nil
}
