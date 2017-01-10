package httpService

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/encrypt"
	"github.com/bparli/goavail/ipState"
	"github.com/gorilla/mux"
)

//HTTPUpdate struct to manage IP updates from peer monitoring agents
type HTTPUpdate struct {
	Peer      string
	IPAddress string
	Live      bool
}

func recvNote(w http.ResponseWriter, r *http.Request) {
	update := new(HTTPUpdate)
	err := json.NewDecoder(r.Body).Decode(update)
	if err != nil {
		log.Fatal(err)
	}

	if ipState.Gm.CryptoKey != "" { //payload should be encrypted so need to decrypt
		update.Peer = encrypt.Decrypt([]byte(ipState.Gm.CryptoKey), update.Peer)
		update.IPAddress = encrypt.Decrypt([]byte(ipState.Gm.CryptoKey), update.IPAddress)
		log.Debugln("Payload Decrypted")
	}

	log.Debugln("Received Update: ", update.IPAddress, update.Live, update.Peer)
	liveCheck := ipState.UpdateGlobalState(update.IPAddress, update.Live, update.Peer) //update global state and check overall status
	if liveCheck >= ipState.Gm.MinAgreement || liveCheck <= -1*ipState.Gm.MinAgreement {
		log.Debugln("Have received enough Peer agreement.  Received ", liveCheck, " agreements, and need ", ipState.Gm.MinAgreement)
		ipState.NotifyIPState(update.IPAddress, update.Live, true)
	} else {
		log.Debugln("Have not received enough Peer agreement.  Received ", liveCheck, " agreements, but need ", ipState.Gm.MinAgreement)
	}
	if err != nil {
		log.Debugln(err)
	}
}

//UpdatesListener listens for updates from peers
func UpdatesListener(localAddr string) {
	router := mux.NewRouter()
	router.HandleFunc("/", recvNote)
	http.ListenAndServe(localAddr, router)
}
