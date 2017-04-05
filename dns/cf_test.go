package dns

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Mock DNS
type mockCf struct {
}

func (d *mockCf) ZoneIDByName(domain string) (int, error) {
	return 123456, nil
}

func (d *mockCf) DNSRecords(name string, ipAddress string, dryRun bool) ([]string, error) {
	return []string{"Something Already in Here"}, nil
}

func Init() {
	// Make sure we can't possibly do damage
	os.Unsetenv("CF_API_KEY")
	os.Unsetenv("CF_API_EMAIL")
}

func Test_AddIP(t *testing.T) {
	Convey("AddIp() should fail", t, func() {
		cf := ConfigureCloudflare("testing.com", true, []string{"127.0.0.1", "127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})
		err := cf.AddIP("127.0.0.1", true)
		So(err, ShouldNotEqual, nil)
	})
}

func Test_RemoveIP(t *testing.T) {
	Convey("RemoveIp() should fail", t, func() {
		cf := ConfigureCloudflare("testing.com", true, []string{"127.0.0.1", "127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})
		err := cf.RemoveIP("127.0.0.1", true)
		So(err, ShouldNotEqual, nil)
	})
}

func Test_ConfigureCloudFlare(t *testing.T) {
	Convey("ConfigureCloudFlare()", t, func() {
		cf := ConfigureCloudflare("testing.com", true, []string{"127.0.0.1", "127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})

		So(cf.DNSDomain, ShouldEqual, "testing.com")
		So(cf.Proxied, ShouldEqual, true)
		So(cf.Addresses[0], ShouldEqual, "127.0.0.1")
		So(cf.Addresses[1], ShouldEqual, "127.0.0.2")
		So(cf.Hostnames[0], ShouldEqual, "beavis.dummy.com")
		So(cf.Hostnames[1], ShouldEqual, "butthead.dummy.com")
	})
}

func Test_FormatHostnameCF(t *testing.T) {
	Convey("formatHostname()", t, func() {
		cf := ConfigureCloudflare("testing.com", true, []string{"127.0.0.1"}, []string{"beavis.dummy.com"})

		Convey("returns the right string when not including a domain", func() {
			So(cf.formatHostname("shakespeare"), ShouldEqual, "shakespeare.testing.com")
		})

		Convey("does not append an extra domain name", func() {
			So(cf.formatHostname("shakespeare.testing.com"), ShouldEqual, "shakespeare.testing.com")
		})
	})
}
