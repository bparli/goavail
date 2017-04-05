package httpService

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/encrypt"
	checks "github.com/bparli/goavail/health_checks"
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

	if checks.Gm.CryptoKey != "" { //payload should be encrypted so need to decrypt
		update.Peer = encrypt.Decrypt([]byte(checks.Gm.CryptoKey), update.Peer)
		update.IPAddress = encrypt.Decrypt([]byte(checks.Gm.CryptoKey), update.IPAddress)
		log.Debugln("Payload Decrypted")
	}

	log.Debugln("Received Update: ", update.IPAddress, update.Live, update.Peer)
	liveCheck := checks.UpdateGlobalState(update.IPAddress, update.Live, update.Peer) //update global state and check overall status
	if liveCheck >= checks.Gm.MinAgreement || liveCheck <= -1*checks.Gm.MinAgreement {
		log.Debugln("Have received enough Peer agreement.  Received ", liveCheck, " agreements, and need ", checks.Gm.MinAgreement)
		checks.NotifyIPState(update.IPAddress, update.Live, true)
	} else {
		log.Debugln("Have not received enough Peer agreement.  Received ", liveCheck, " agreements, but need ", checks.Gm.MinAgreement)
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
