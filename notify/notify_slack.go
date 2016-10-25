package notify

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

type slackNotifications struct {
	UseSlack  bool
	slackAddr string
}

var SlackNotify slackNotifications

type Params map[string]string

func InitSlack(slackAddr string) {
	if slackAddr != "" {
		SlackNotify.UseSlack = true
		SlackNotify.slackAddr = slackAddr
	}
}

func (s *slackNotifications) SendToSlack(ipAddr string, domain string, addOrRemove string, dryRun bool) {
	log.Debugln("Dry Run is True, sending Slack notification", "https://hooks.slack.com/services/"+s.slackAddr)
	text := ""
	if dryRun {
		text = "Dry Run set to True.  Agent would have " + addOrRemove + " " + ipAddr + " and configured entries from/to domain " + domain + " "
	} else {
		text = "Dry Run set to False. Agent " + addOrRemove + " " + ipAddr + " and configured entries from/to domain " + domain + " "
	}
	var p Params = map[string]string{
		"text": text,
	}
	json, _ := json.Marshal(p)
	resp, _ := http.PostForm("https://hooks.slack.com/services/"+s.slackAddr, url.Values{"payload": []string{string(json)}})
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Errorf("status code: %d, response body: %s", resp.StatusCode, body)
	}

}
