package ipState

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tatsushid/go-fastping"
)

func Init() {
	// Make sure we can't possibly do damage
	os.Unsetenv("CF_API_KEY")
	os.Unsetenv("CF_API_EMAIL")
}

type MockCFlare struct {
	DNSDomain string
	Proxied   bool
	Addresses []string
	Hostnames []string
}

type MockDNSProvider interface {
	MockAddIp(ipAddress string) error
	MockRemoveIp(ipAddress string) error
}

func (r *MockCFlare) AddIP(ipAddress string, dryRun bool) error {
	DNSCount++
	return nil
}

func (r *MockCFlare) RemoveIP(ipAddress string, dryRun bool) error {
	DNSCount--
	return nil
}

func (r *MockCFlare) addDNSName(name string, ipAddress string) error {
	return nil
}

func (r *MockCFlare) deleteDNSName(name string, ipAddress string) error {
	return nil
}

func (r *MockCFlare) formatHostname(host string) string {
	if strings.Contains(host, r.DNSDomain) {
		return host
	}
	return host + "." + r.DNSDomain
}

func mockInitMaster(dnsConfig *MockCFlare, threshold int) {
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
		Master.AddressSuccesses[ip] = 4 //initialize IPs such that they are already in service at start time
	}

	Master.P.MaxRTT = 2 * time.Second
}

var DNSCount = 0

func mockConfigureCloudflare() (*MockCFlare, error) {
	adds := []string{"52.52.52.52"}
	host := []string{"dummy_host"}
	dnsConfig := MockCFlare{
		"dummy_domain",
		true,
		adds,
		host}

	return &dnsConfig, nil
}

func Test_handleTransition(t *testing.T) {
	Convey("When running fastping", t, func() {
		dnsConfig, _ := mockConfigureCloudflare()
		InitGM(dnsConfig.Addresses, true)
		Gm.Clustered = false
		mockInitMaster(dnsConfig, 3)
		handleTransition("52.52.52.52", true)
		So(DNSCount, ShouldEqual, 1)
		handleTransition("52.52.52.52", false)
		So(DNSCount, ShouldEqual, 0)
	})

}
