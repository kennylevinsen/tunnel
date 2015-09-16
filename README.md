# tunnel
Local transport proxy. Runs a local server, and proxies it with transport
handling to the provided destination. An example usage would be to connect to a
serve2 endpoint to have a stealthy TLS bridge, running a VPN or SSH over it.

Usage:

	tunnel [options] [local transport:host:port] [remote transport:host:port]

Example:

	tunnel plain::9999 tls:example.com:443
	ssh -p 9999 localhost

... or:

	ssh -oProxyCommand="tunnel - tls:%h:%p" -p443 example.com

Installation

	go get github.com/joushou/tunnel

For more information on serve2 and serve2d, see
http://github.com/joushou/serve2d and http://github.com/joushou/serve2
