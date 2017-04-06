package checks

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/bparli/goavail/encrypt"
	"github.com/jarcoal/httpmock"
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

func Test_PeerNotifications(t *testing.T) {
	Convey("Test Peer Notification", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "http://some.dummy.service:7979",
			func(req *http.Request) (*http.Response, error) {
				update := new(HTTPUpdate)
				err := json.NewDecoder(req.Body).Decode(update)
				if err != nil {
					log.Fatal(err)
				}
				ipAddr := encrypt.Decrypt([]byte("example key 1234"), update.IPAddress)
				So(ipAddr, ShouldEqual, "192.192.168.168")
				So(update.Live, ShouldEqual, true)

				resp, err := httpmock.NewJsonResponse(200, nil)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		InitGM([]string{"192.192.168.168"}, true)
		Gm.Peers = []string{"some.dummy.service:7979"}
		Gm.CryptoKey = "example key 1234"
		err := notifyPeers("192.192.168.168", true)
		So(err, ShouldEqual, nil)
	})
}
