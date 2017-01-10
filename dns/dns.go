package dns

//Provider interface for DNS service provider interface
type Provider interface {
	AddIP(ipAddress string, dryRun bool) error
	RemoveIP(ipAddress string, dryRun bool) error
}
