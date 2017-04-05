package checks

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_InitGm(t *testing.T) {
	Convey("Init the Global Map", t, func() {
		adds := []string{"192.168.1.10:80", "192.168.1.11:80"}
		InitGM(adds, true)
		So(Gm.IPLive["192.168.1.10:80"], ShouldEqual, true)
		So(Gm.IPLive["192.168.1.11:80"], ShouldEqual, true)
	})
}

func Test_NotifyIpState(t *testing.T) {
	Convey("Init the Global Map", t, func() {
		dnsConfig, _ := mockConfigureCloudflare()
		InitGM(dnsConfig.Addresses, true)
		Gm.Clustered = false
		NotifyIPState("192.168.1.10", true, false)
		So(Gm.IPLive["192.168.1.10"], ShouldEqual, true)
		NotifyIPState("192.168.1.10", true, true)
		So(Gm.IPLive["192.168.1.10"], ShouldEqual, true)

		Gm.IPLive["192.168.1.10"] = false
		NotifyIPState("192.168.1.10", false, true)
		So(Gm.IPLive["192.168.1.10"], ShouldEqual, false)
	})
}

func Test_initPeersIpViews(t *testing.T) {
	Convey("Init the Global Map Peers IP States", t, func() {
		adds := []string{"52.52.52.52", "53.53.53.53"}
		InitGM(adds, true)
		Gm.Peers = []string{"192.168.1.10:80", "192.168.1.11:80"}
		InitPeersIPViews()
		Gm.Clustered = true
		So(Gm.peersIPView["52.52.52.52"].IpsView["192.168.1.10:80"], ShouldEqual, true)
		So(Gm.peersIPView["53.53.53.53"].IpsView["192.168.1.10:80"], ShouldEqual, true)
		So(Gm.peersIPView["52.52.52.52"].IpsView["192.168.1.11:80"], ShouldEqual, true)
		So(Gm.peersIPView["53.53.53.53"].IpsView["192.168.1.11:80"], ShouldEqual, true)
	})
}

func Test_UpdateGlobalState(t *testing.T) {
	Convey("Init the Global Map Peers IP States", t, func() {
		adds := []string{"52.52.52.52", "53.53.53.53"}
		InitGM(adds, true)
		Gm.Peers = []string{"192.168.1.10:80", "192.168.1.11:80"}
		InitPeersIPViews()
		Gm.Clustered = true
		liveCheck := UpdateGlobalState("52.52.52.52", true, "192.168.1.10:80")
		So(liveCheck, ShouldEqual, 2)

		//Gm.peersIpView["52.52.52.52"].IpsView["192.168.1.10:80"] = false
		liveCheck = UpdateGlobalState("52.52.52.52", false, "192.168.1.10:80")
		So(liveCheck, ShouldEqual, 0)
	})
}
