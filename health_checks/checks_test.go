package checks

import (
	"net"
	"testing"

	"github.com/bparli/goavail/dns"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_newChecks(t *testing.T) {
	Convey("Init the Health Checks Struct", t, func() {
		r53 := dns.ConfigureRoute53("testing", "us-foo-1", 60, "zone-asdfasdf", []string{"127.0.0.1",
			"127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})
		InitGM([]string{"127.0.0.1", "127.0.0.2"}, false)
		Gm.Type = "tcp"

		NewChecks(r53, 5, 10, 100)
		So(Master.DNS, ShouldEqual, r53)
		So(Master.Results["127.0.0.1"], ShouldEqual, nil)
		So(Master.AddressFails["127.0.0.1"], ShouldEqual, 0)
		So(Master.AddressSuccesses["127.0.0.1"], ShouldEqual, 6)
	})
}

func Test_tcpChecks(t *testing.T) {
	Convey("Verify TCP Health Checks", t, func() {
		go func() {
			l, err := net.Listen("tcp", ":3000")
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			for {
				conn, err := l.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()

		r53 := dns.ConfigureRoute53("testing", "us-foo-1", 60, "zone-asdfasdf", []string{"127.0.0.1"},
			[]string{"beavis.dummy.com", "butthead.dummy.com"})
		InitGM([]string{"127.0.0.1"}, true)
		Gm.Type = "tcp"

		NewChecks(r53, 5, 10, 3000)
		onSuccess, onFail := make(chan *response), make(chan *response)
		go runTCPChecks("127.0.0.1:3000", onSuccess, onFail)
		res := <-onSuccess
		So(res.tcpAddr, ShouldEqual, "127.0.0.1:3000")

		NewChecks(r53, 5, 10, 4000)
		onSuccess2, onFail2 := make(chan *response), make(chan *response)
		go runTCPChecks("127.0.0.1:4000", onSuccess2, onFail2)
		res = <-onFail2
		So(res.tcpAddr, ShouldEqual, "127.0.0.1:4000")
	})
}
