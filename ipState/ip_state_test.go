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

func Test_initPeersIpViews(t *testing.T) {
	Convey("Init the Global Map Peers IP States", t, func() {
		adds := []string{"52.52.52.52", "53.53.53.53"}
		InitGM(adds, true)
		Gm.Peers = []string{"192.168.1.10:80", "192.168.1.11:80"}
		InitPeersIpViews()
		Gm.Clustered = true
		So(Gm.peersIpView["52.52.52.52"].IpsView["192.168.1.10:80"], ShouldEqual, true)
		So(Gm.peersIpView["53.53.53.53"].IpsView["192.168.1.10:80"], ShouldEqual, true)
		So(Gm.peersIpView["52.52.52.52"].IpsView["192.168.1.11:80"], ShouldEqual, true)
		So(Gm.peersIpView["53.53.53.53"].IpsView["192.168.1.11:80"], ShouldEqual, true)
	})
}
