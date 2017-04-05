package checks

import (
	"testing"

	"github.com/bparli/goavail/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_initMaster(t *testing.T) {
	Convey("Init the Pingmon Master", t, func() {
		dnsConfig := dns.ConfigureCloudflare("testing.com", true, []string{"127.0.0.1", "127.0.0.2"},
			[]string{"beavis.dummy.com", "butthead.dummy.com"})
		InitGM([]string{"127.0.0.1", "127.0.0.2"}, false)
		NewChecks(dnsConfig, 5, 10, nil)
		So(Master.DNS, ShouldEqual, dnsConfig)
		So(Master.Results["127.0.0.1"], ShouldEqual, nil)
		So(Master.AddressFails["127.0.0.1"], ShouldEqual, 0)
		So(Master.AddressSuccesses["127.0.0.1"], ShouldEqual, 6)
	})
}
