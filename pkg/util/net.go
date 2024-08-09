package util

import (
	"net"
	"net/http"
	"strings"
)

// Extracts and returns the clients IP from the given request.
// Looks at X-Forwarded-For header, X-Real-Ip header and request.RemoteAddr in that order.
// Returns nil if none of them are set or is set to an invalid value.
func GetClientIP(req *http.Request) net.IP {
	ips := SourceIPs(req)
	if len(ips) == 0 {
		return nil
	}
	return ips[0]
}

// SourceIPs splits the comma separated X-Forwarded-For header and joins it with
// the X-Real-Ip header and/or req.RemoteAddr, ignoring invalid IPs.
// The X-Real-Ip is omitted if it's already present in the X-Forwarded-For chain.
// The req.RemoteAddr is always the last IP in the returned list.
// It returns nil if all of these are empty or invalid.
func SourceIPs(req *http.Request) []net.IP {
	var srcIPs []net.IP

	hdr := req.Header
	// First check the X-Forwarded-For header for requests via proxy.
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrForwardedFor != "" {
		// X-Forwarded-For can be a csv of IPs in case of multiple proxies.
		// Use the first valid one.
		parts := strings.Split(hdrForwardedFor, ",")
		for _, part := range parts {
			ip := net.ParseIP(strings.TrimSpace(part))
			if ip != nil {
				srcIPs = append(srcIPs, ip)
			}
		}
	}

	// Try the X-Real-Ip header.
	hdrRealIp := hdr.Get("X-Real-Ip")
	if hdrRealIp != "" {
		ip := net.ParseIP(hdrRealIp)
		// Only append the X-Real-Ip if it's not already contained in the X-Forwarded-For chain.
		if ip != nil && !containsIP(srcIPs, ip) {
			srcIPs = append(srcIPs, ip)
		}
	}

	// Always include the request Remote Address as it cannot be easily spoofed.
	var remoteIP net.IP
	// Remote Address in Go's HTTP server is in the form host:port so we need to split that first.
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		remoteIP = net.ParseIP(host)
	}
	// Fallback if Remote Address was just IP.
	if remoteIP == nil {
		remoteIP = net.ParseIP(req.RemoteAddr)
	}

	// Don't duplicate remote IP if it's already the last address in the chain.
	if remoteIP != nil && (len(srcIPs) == 0 || !remoteIP.Equal(srcIPs[len(srcIPs)-1])) {
		srcIPs = append(srcIPs, remoteIP)
	}

	return srcIPs
}

// Checks whether the given IP address is contained in the list of IPs.
func containsIP(ips []net.IP, ip net.IP) bool {
	for _, v := range ips {
		if v.Equal(ip) {
			return true
		}
	}
	return false
}
