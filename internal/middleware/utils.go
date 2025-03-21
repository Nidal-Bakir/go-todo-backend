package middleware

import (
	"net"
	"net/http"
)

type Middleware func(http.Handler) http.HandlerFunc

// Wrappe an endpoint with a chain of middlewares
//
// e.g:
//
//	middleware.MiddlewareChain(
//		CreateNewBook, // <-- wrapped by the next middlewares:
//		middleware.ACT_app_x_www_form_urlencoded, // <-- First middleware to be run
//		middleware.Cors, // <-- Second
//		middleware.Auth, // <-- Third
//		middleware.RateLimiter, // <-- Fourth, and last
//	)
func MiddlewareChain(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) < 1 {
		return h
	}

	wrapped := h

	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped
}

// canonicalizeIP returns a form of ip suitable for comparison to other IPs.
// For IPv4 addresses, this is simply the whole string.
// For IPv6 addresses, this is the /64 prefix.
func canonicalizeIP(ip string) string {
	isIPv6 := false
	// This is how net.ParseIP decides if an address is IPv6
	// https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/net/ip.go;l=704
	for i := 0; !isIPv6 && i < len(ip); i++ {
		switch ip[i] {
		case '.':
			// IPv4
			return ip
		case ':':
			// IPv6
			isIPv6 = true
			break
		}
	}

	if !isIPv6 {
		// Not an IP address at all
		return ip
	}

	ipv6 := net.ParseIP(ip)
	if ipv6 == nil {
		return ip
	}

	return ipv6.Mask(net.CIDRMask(64, 128)).String()
}
