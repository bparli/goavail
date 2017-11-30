package dns

import (
	"os"
	"strings"

	"github.com/bparli/goavail/notify"
	cloudflare "github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

//CloudFlare to maintain cloudflare addresses and domain in scope
type CloudFlare struct {
	DNSDomain string   `toml:"dns_domain"`
	Proxied   bool     `toml:"dns_proxied"`
	Addresses []string `toml:"ip_addresses"`
	Hostnames []string `toml:"hostnames"`
}

//GetAddrs returns the addresses in scope for monitoring purposes
func (r *CloudFlare) GetAddrs() []string {
	return r.Addresses
}

//ConfigureCloudflare to initialize CloudFlare struct
func ConfigureCloudflare(domain string, proxied bool, addresses []string, hostnames []string) *CloudFlare {
	log.Debugln("Addresses and Hostnames:", addresses, hostnames)
	return &CloudFlare{domain, proxied, addresses, hostnames}
}

func (r *CloudFlare) formatHostname(host string) string {
	if strings.Contains(host, r.DNSDomain) {
		return host
	}
	return host + "." + r.DNSDomain
}

//AddIP to add IP back into cloudflare domain
func (r *CloudFlare) AddIP(ipAddress string, dryRun bool) error {
	for _, name := range r.Hostnames {
		err := r.addDNSName(name, ipAddress, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CloudFlare) addDNSName(name string, ipAddress string, dryRun bool) error {
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
func (r *CloudFlare) RemoveIP(ipAddress string, dryRun bool) error {
	for _, name := range r.Hostnames {
		err := r.deleteDNSName(name, ipAddress, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CloudFlare) deleteDNSName(name string, ipAddress string, dryRun bool) error {
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
