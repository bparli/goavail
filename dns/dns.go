package dns

type DnsProvider interface {
	AddIp(ipAddress string, dryRun bool) error
	RemoveIp(ipAddress string, dryRun bool) error
}
