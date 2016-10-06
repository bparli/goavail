package httpService

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/ipState"
	"github.com/gorilla/mux"
)

type HttpUpdate struct {
	Peer      string
	IpAddress string
	Live      bool
}

func recvNote(w http.ResponseWriter, r *http.Request) {
	update := new(HttpUpdate)
	err := json.NewDecoder(r.Body).Decode(update)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Received Update: ", update.IpAddress, update.Live, update.Peer)
	ipState.NotifyIpState(update.IpAddress, update.Live, true)
	if err != nil {
		log.Debugln(err)
	}
}

//UpdateListener listens for updates from peers
func UpdatesListener(localAddr string) {
	router := mux.NewRouter()
	router.HandleFunc("/", recvNote)
	http.ListenAndServe(localAddr, router)
}
