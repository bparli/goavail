package dns

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

//Route53 to maintain Route53 addresses and domain in scope
type Route53 struct {
	DNSDomain string   `toml:"aws_domain"`
	AWSRegion string   `toml:"aws_region"`
	TTL       int64    `toml:"ttl"`
	AWSZoneID string   `toml:"aws_zoneid"`
	Addresses []string `toml:"ip_addresses"`
	Hostnames []string `toml:"hostnames"`
}

//ConfigureRoute53 object
func ConfigureRoute53(domain string, region string, ttl int64, zoneID string, addresses []string, hostnames []string) *Route53 {
	return &Route53{domain, region, ttl, zoneID, addresses, hostnames}
}

func (r *Route53) GetAddrs() []string {
	return r.Addresses
}

func (r *Route53) GetRecords(hostname string) (*route53.ListResourceRecordSetsOutput, error) {
	svc := route53.New(session.New(), aws.NewConfig().WithRegion(r.AWSRegion))

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(r.AWSZoneID), // Required
		StartRecordName: aws.String(hostname + "." + r.DNSDomain),
		StartRecordType: aws.String("A"),
	}
	resp, err := svc.ListResourceRecordSets(params)

	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return resp, nil
}

//AddIP to add IP back into Route53 zone
func (r *Route53) AddIP(ipAddress string, dryRun bool) error {
	log.Debugln("DNS Configs: ", r.DNSDomain, r.Hostnames, r.Addresses)
	if dryRun {
		log.Infof(
			"DNS: Would have ADDED IP address '%s' but it's a DRY RUN", ipAddress)
		return nil
	}
	for _, name := range r.Hostnames {
		err := r.runChange(name, ipAddress, "ADD")
		if err != nil {
			return err
		}
	}
	return nil
}

//RemoveIP to remove IP back from Route53 zone
func (r *Route53) RemoveIP(ipAddress string, dryRun bool) error {
	if dryRun {
		log.Infof(
			"DNS: Would have DELETED IP address '%s' but it's a DRY RUN", ipAddress)
		return nil
	}
	for _, name := range r.Hostnames {
		err := r.runChange(name, ipAddress, "DELETE")
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Route53) runChange(name string, ipAddress string, changeType string) error {
	request, err := r.buildDNSChangeRequest(name, ipAddress, changeType)
	if err != nil {
		return err
	}
	if request != nil {
		return r.modifyDNSRecord(request)
	}
	return nil
}

func (r *Route53) formatHostname(host string) string {
	if strings.Contains(host, r.DNSDomain) {
		return host
	}

	return host + "." + r.DNSDomain
}

func (r *Route53) buildDNSChangeRequest(name string, ipAddress string, changeType string) (*route53.ChangeResourceRecordSetsInput, error) {
	// Make sure we only use short names here
	fields := strings.Split(name, ".")
	name = fields[0]

	var recordSet *route53.ResourceRecordSet
	setID := name

	currRecords, err := r.GetRecords(name)
	if err != nil {
		return nil, err
	}
	var resRecords []*route53.ResourceRecord
	if changeType == "ADD" { //Adding Ip address
		for _, rec := range currRecords.ResourceRecordSets[0].ResourceRecords {
			if *rec.Value == ipAddress {
				log.Debugln("Address is already added, do nothing")
				return nil, nil
			}
			resRecords = append(resRecords, &route53.ResourceRecord{Value: rec.Value})
		}
		resRecords = append(resRecords, &route53.ResourceRecord{Value: aws.String(ipAddress)})
	} else { //Removing IP Address
		addrMissing := true
		for _, rec := range currRecords.ResourceRecordSets[0].ResourceRecords {
			if *rec.Value == ipAddress {
				addrMissing = false
				continue
			}
			resRecords = append(resRecords, &route53.ResourceRecord{Value: rec.Value})
		}
		if addrMissing {
			log.Debugln("Address is already removed, do nothing")
			return nil, nil
		}
	}
	recordSet = &route53.ResourceRecordSet{
		Name:            aws.String(r.formatHostname(name)),
		Type:            aws.String("A"),
		ResourceRecords: resRecords,
		SetIdentifier:   &setID,
		TTL:             &r.TTL,
		Weight:          aws.Int64(0),
	}

	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(r.AWSZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action:            aws.String("UPSERT"),
					ResourceRecordSet: recordSet,
				},
			},
		},
	}

	err = params.Validate()
	if err != nil {
		return nil, err
	}
	log.Debugln(params)
	return params, nil
}

func (r *Route53) modifyDNSRecord(request *route53.ChangeResourceRecordSetsInput) error {
	svc := route53.New(session.New(), aws.NewConfig().WithRegion(r.AWSRegion))
	_, err := svc.ChangeResourceRecordSets(request)
	if err != nil {
		return err
	}

	return nil
}

var predefinedHostedZonesIds = map[string]string{
	"ap-northeast-1": "Z14GRHDCWA56QT",
	"ap-northeast-2": "ZWKZPGTI48KDX",
	"ap-south-1":     "ZP97RAFLXTNZK",
	"ap-southeast-1": "Z1LMS91P8CMLE5",
	"ap-southeast-2": "Z1GM3OXH4ZPM65",
	"ca-central-1":   "ZQSVJUPU6J1EY",
	"eu-central-1":   "Z215JYRZR1TBD5",
	"eu-west-1":      "Z32O12XQLNTSW2",
	"eu-west-2":      "ZHURV8PSTC4K8",
	"us-east-1":      "Z35SXDOTRQ7X7K",
	"us-east-2":      "Z3AADJGX6KTTL2",
	"us-west-1":      "Z368ELLRRE2KJ0",
	"us-west-2":      "Z1H1FL5HABSF5",
	"sa-east-1":      "Z2P70J7HTTTPLU",
}

// see https://forums.aws.amazon.com/thread.jspa?messageID=608949
func (r *Route53) getHostedZoneID() string {
	return predefinedHostedZonesIds[r.AWSRegion]
}
