package ipState

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_InitGm(t *testing.T) {
	Convey("Init the Global Map", t, func() {
		adds := []string{"192.168.1.10:80", "192.168.1.11:80"}
		InitGM(adds, true)
		So(Gm.IpLive["192.168.1.10:80"], ShouldEqual, true)
		So(Gm.IpLive["192.168.1.11:80"], ShouldEqual, true)
	})
}

func Test_NotifyIpState(t *testing.T) {
	Convey("Init the Global Map", t, func() {
		dnsConfig, _ := mockConfigureCloudflare()
		InitGM(dnsConfig.Addresses, true)
		Gm.Clustered = false
		NotifyIpState("192.168.1.10", true, false)
		So(Gm.IpLive["192.168.1.10"], ShouldEqual, true)
	})
}
