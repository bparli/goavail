package dns

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewRoute53(t *testing.T) {
	Convey("NewRoute53()", t, func() {
		Convey("returns a properly configured Route53 instance", func() {
			r53 := ConfigureRoute53("testing", "us-foo-1", 60, "zone-asdfasdf", []string{"127.0.0.1", "127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})

			So(r53.DNSDomain, ShouldEqual, "testing")
			So(r53.AWSRegion, ShouldEqual, "us-foo-1")
			So(r53.AWSZoneID, ShouldEqual, "zone-asdfasdf")
		})
	})
}

func TestDryRun(t *testing.T) {
	Convey("When in Dry Run Mode", t, func() {
		r53 := ConfigureRoute53("example.com", "us-foo-1", 60, "zone-asdfasdf", []string{"127.0.0.1", "127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})

		Convey("we can add a DNS name", func() {
			result := r53.AddIP("172.16.16.1", true)

			So(result, ShouldBeNil)
		})

		Convey("we can delete a DNS record", func() {
			result := r53.RemoveIP("172.16.16.1", true)

			So(result, ShouldBeNil)
		})
	})
}

func TestFormatHostname(t *testing.T) {
	Convey("formatHostname()", t, func() {
		r53 := ConfigureRoute53("example.com", "us-foo-1", 60, "zone-asdfasdf", []string{"127.0.0.1", "127.0.0.2"}, []string{"beavis.dummy.com", "butthead.dummy.com"})

		Convey("returns the right string when not including a domain", func() {
			So(r53.formatHostname("shakespeare"), ShouldEqual, "shakespeare.example.com")
		})

		Convey("does not append an extra domain name", func() {
			So(r53.formatHostname("shakespeare.example.com"), ShouldEqual, "shakespeare.example.com")
		})
	})
}
