package dns

type DnsProvider interface {
	AddIp(ipAddress string) error
	RemoveIp(ipAddress string) error
}
