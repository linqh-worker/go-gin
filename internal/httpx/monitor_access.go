package httpx

import (
	"net"
	"net/http"
	"strings"
)

// MonitorAllowed 仅允许白名单直连；监控不接受代理转发，避免将代理IP误认为访问者IP。
func MonitorAllowed(r *http.Request, whitelist []string) bool {
	if r == nil {
		return false
	}
	for _, header := range []string{"Forwarded", "X-Forwarded-For", "X-Real-IP"} {
		if strings.TrimSpace(r.Header.Get(header)) != "" {
			return false
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, allowed := range whitelist {
		candidate := net.ParseIP(strings.TrimSpace(allowed))
		if candidate != nil && ip.Equal(candidate) {
			return true
		}
	}
	return false
}
