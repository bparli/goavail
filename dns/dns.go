package dns

type DnsProvider interface {
	AddIP(ipAddress string) error
	RemoveIP(ipAddress string) error
}
