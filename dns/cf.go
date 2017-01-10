package dns

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bparli/goavail/notify"
	"github.com/cloudflare/cloudflare-go"
)

//CFlare to maintain cloudflare addresses and domain in scope
type CFlare struct {
	DNSDomain string
	Proxied   bool
	Addresses []string
	Hostnames []string
}

//ConfigureCloudflare to initialize CFlare struct
func ConfigureCloudflare(domain string, proxied bool, addresses []string, hostnames []string) (*CFlare, error) {
	dnsConfig := CFlare{
		domain,
		proxied,
		addresses,
		hostnames}

	return &dnsConfig, nil
}

func (r *CFlare) formatHostname(host string) string {
	if strings.Contains(host, r.DNSDomain) {
		return host
	}
	return host + "." + r.DNSDomain
}

//AddIP to add IP back into cloudflare domain
func (r *CFlare) AddIP(ipAddress string, dryRun bool) error {
	for _, name := range r.Hostnames {
		err := r.addDNSName(name, ipAddress, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CFlare) addDNSName(name string, ipAddress string, dryRun bool) error {
	// Construct a new API object
	log.Infoln("Adding", name, ipAddress, "to Cloudflare")
	api, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	if err != nil {
		return err
	}

	// Fetch the zone ID
	zoneID, err := api.ZoneIDByName(r.DNSDomain) // Assumes exists in CloudFlare already
	if err != nil {
		return err
	}

	params := &cloudflare.DNSRecord{
		Type:     "A",
		Name:     r.formatHostname(name),
		ZoneID:   zoneID,
		ZoneName: r.DNSDomain,
		Content:  ipAddress,
		Proxied:  r.Proxied,
	}

	dnsRec, err := api.DNSRecords(zoneID, *params)
	if err != nil {
		return err
	}
	if len(dnsRec) == 1 {
		log.Infoln("DNS Record already added")
		return nil
	}
	if dryRun {
		log.Infoln("Dry Run is True.  Would have updated DNS for address " + ipAddress)
	} else {
		resp, err := api.CreateDNSRecord(zoneID, *params)
		if err != nil {
			return err
		}
		log.Debugln("CF response", resp)
	}
	if notify.SlackNotify.UseSlack == true {
		notify.SlackNotify.SendToSlack(ipAddress, r.DNSDomain, "Added", dryRun)
	}
	return nil
}

//RemoveIP to remove IP from Cloudflare domain
func (r *CFlare) RemoveIP(ipAddress string, dryRun bool) error {
	for _, name := range r.Hostnames {
		err := r.deleteDNSName(name, ipAddress, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CFlare) deleteDNSName(name string, ipAddress string, dryRun bool) error {
	// Construct a new API object
	log.Infoln("Deleting", name, ipAddress, "from Cloudflare")
	api, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	if err != nil {
		return err
	}

	// Fetch the zone ID
	zoneID, err := api.ZoneIDByName(r.DNSDomain) // Assumes exists in CloudFlare already
	if err != nil {
		return err
	}
	log.Debugln(r.Proxied, ipAddress, r.DNSDomain, zoneID, r.formatHostname(name))
	params := &cloudflare.DNSRecord{
		Type:     "A",
		Name:     r.formatHostname(name),
		ZoneID:   zoneID,
		ZoneName: r.DNSDomain,
		Content:  ipAddress,
		Proxied:  r.Proxied,
	}

	dnsRec, err := api.DNSRecords(zoneID, *params)
	if err != nil {
		return err
	}

	log.Infoln(dnsRec)

	if len(dnsRec) == 0 {
		log.Infoln("DNS Record already removed")
		return nil
	}
	if dryRun {
		log.Infoln("Dry Run is True.  Would have updated DNS for address " + ipAddress)
	} else {
		err = api.DeleteDNSRecord(zoneID, dnsRec[0].ID)
		if err != nil {
			return err
		}
	}
	if notify.SlackNotify.UseSlack == true {
		notify.SlackNotify.SendToSlack(ipAddress, r.DNSDomain, "Removed", dryRun)
	}
	return nil
}
