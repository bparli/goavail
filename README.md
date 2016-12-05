#Goavail
===============
Goavail is an IP monitoring and fast DNS failover agent written in Go.  

This personal project was done as part of Nitro Software's hack week.  We run our systems across multiple AWS Availability Zones, but are still succeptible if an AZ itself goes down.  This tool is meant to serve as an external monitoring agent, able to trigger DNS A record updates should an AWS zone go down.  The agent continuously pings a set of IP Addresses (EIPs) in a certain domain, noting their health.  After three consecutive misses, the IP and its associated A records are removed from DNS.  The agent will continue to monitor the EIP and, should it come back alive for some reason, will enter it back into service with the appropriate API call via the DNS interface.  

These EIPs live in separate AZs, and since we're also proxying these EIPs through our CDN/DNS, this failover should be a matter of seconds.

##Running
To run the agent first clone and build with Godep.  Then run the binary with the "monitor" command and specify the configuration file (default is goavail.toml).  The agent runs in "dry-run" mode by default.  Dry-run mode is totally passive and no DNS records can be updated.  For example: 

```bash
$ godep go build
$ ./goavail monitor --config-file --no-dry-run goavail.toml
```
The agent must be run as root user since its using raw socket for the ICMP.  Also see goavail.toml for a configration example

##Cluster Mode
Cluster mode uses a fork of the [memberlist library] (https://github.com/hashicorp/memberlist) to verify peers' health.  If a node sees an IP failure, it will notify its peers before invoking the DNS interface to take the failed IP out of the Pool.  A node will only remove an IP from the pool once its noted a failure and recieved confirmation from a peer.  

To run in Cluster mode simply add the "peers" and "local_addr" entries in the toml file.  

Since running in cluster mode will likely mean the agents are running over the WAN, it makes sense to encrypt some of the sensitive updates communication between peers.  To encrypt the payload between peers, simply add the `crypto_key` setting in the configuration file with the private key shared between all the agents.

The `min_peers_agree` setting is the minimum number of peers updates that must be received before an agent can perform an update on an A record.  So if `min_peers_agree = 1` and agent A sees an EIP is down, it must wait until it receives an update from at least one other peer also noting that same EIP as down.  Only then, can agent A take the EIP out of service.

##Configuration File Settings
See goavail.toml for example settings
* __ip_addresses__: The public IP addresses to be monitored
* __failure_threshold__: Sensitivity to ping failures
* __dns_domain__: the domain to be monitored
* __hostnames__: The hostnames to be monitored within the above domain.  For example, if the domain was gonitro.com and the hostname cloud, we would be monitoring cloud.gonitro.com
* __dns_proxied__: this is a Cloudflare setting we generally use
* __peers__: list of peer monitoring agents to communicate with over the WAN
* __local_addr__: the local IP address to advertise to the peers
* __members_port__: Goavail uses memberlist to ensure the peers are online.  This allows a custom setting for the memberlist port
* __min_peers_agree__: if the agent detects a change (and the above failure_threshold), it will notify its peers.  It must receive agreement from at least min_peers_agree before the agent can take any action.  This is to reduce false positive likelihood.
* __crypto_key__: optional setting to encrypt the payload in updates to peers.  Note this just sets the AES private key; the IV is hardcoded.

##Reload 
The tool also supports configuration reload.  Simply send a SIGHUP signal to the process and it will reload the modified configuration file
