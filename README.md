#Goavail
===============
Goavail is an IP monitoring and fast DNS failover agent written in Go.  

This personal project was done as part of Nitro Software's hack week.  We run our systems across multiple AWS Availability Zones, but are still succeptible if an AZ itself goes down.  This tool is meant to serve as an external monitoring agent, able to trigger DNS A record updates should an AWS zone go down.  The agent continuously pings a set of IP Addresses (EIPs) in a certain domain, noting their health.  After three consecutive misses, the IP and its associated A records are removed from DNS.  The agent will continue to monitor the EIP and, should it come back alive for some reason, will enter it back into service with the appropriate API call via the DNS interface.  

These EIPs live in separate AZs, and since we're also proxying these EIPs through our CDN/DNS, this failover should be a matter of seconds.

##Running
To run the agent first clone and build with Godep.  Then run the binary with the "monitor" command and specify the configuration file (default is goavail.toml).  For example: `./goavail monitor --config-file goavail.toml` 

The agent must be run as root user since its using raw socket for the ICMP.  Also see goavail.toml for a configration example

##Cluster Mode
Cluster mode uses a fork of the [memberlist library] (https://github.com/hashicorp/memberlist) to verify peers' health.  If a node sees an IP failure, it will notify its peers before invoking the DNS interface to take the failed IP out of the Pool.  A node will only remove an IP from the pool once its noted a failure and recieved confirmation from a peer.  

To run in Cluster mode simply add the "peers" and "local_addr" entries in the toml file.  

Since running in cluster mode will likely mean the agents are running over the WAN, it makes sense to encrypt some of the sensitive updates communication between peers.  To encrypt the payload between peers, simply add the `crypto_key` setting in the configuration file with the private key shared between all the agents.

The `min_peers_agree` setting is the minimum number of peers updates that must be received before an agent can perform an update on an A record.  So if `min_peers_agree = 1` and agent A sees an EIP is down, it must wait until it receives an update from at least one other peer also noting that same EIP as down.  Only then, can agent A take the EIP out of service.

##Reload 
The tool also supports configuration reload.  Simply send a SIGHUP signal to the process and it will reload the modified configuration file
