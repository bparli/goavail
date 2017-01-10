package ipState

import (
	"testing"

	"github.com/bparli/goavail/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_initMaster(t *testing.T) {
	Convey("Init the Pingmon Master", t, func() {
		dnsConfig, _ := dns.ConfigureCloudflare("testing.com", true, []string{"127.0.0.1", "127.0.0.2"},
			[]string{"beavis.dummy.com", "butthead.dummy.com"})
		initMaster(dnsConfig, 5)
		So(Master.DNS, ShouldEqual, dnsConfig)
		So(Master.Results["127.0.0.1"], ShouldEqual, nil)
		So(Master.AddressFails["127.0.0.1"], ShouldEqual, 0)
		So(Master.AddressSuccesses["127.0.0.1"], ShouldEqual, 6)
	})
}
