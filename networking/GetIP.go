package networking

import (
	"net"
	"net/http"
	"strings"
)


func GetIP(r *http.Request) string {
	
	// Common proxy header (may contain multiple, comma-separated IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		// First one is the original client
		return strings.TrimSpace(parts[0])
	}
	
	// Fallback to RemoteAddr (includes port)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	
	return host
}