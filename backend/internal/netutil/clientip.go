// Package netutil содержит общую логику определения IP клиента
// за доверенными прокси (Caddy).
package netutil

import (
	"net"
	"net/http"
	"strings"
)

// ProxyChecker проверяет, является ли адрес доверенным прокси.
// Поддерживает точные IP и CIDR-диапазоны ("127.0.0.1", "10.88.0.0/16").
type ProxyChecker struct {
	ips []net.IP
	nets []*net.IPNet
}

// NewProxyChecker парсит список доверенных прокси один раз при старте.
func NewProxyChecker(trustedProxies []string) *ProxyChecker {
	var ips []net.IP
	var nets []*net.IPNet

	for _, p := range trustedProxies {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "/") {
			if _, ipNet, err := net.ParseCIDR(p); err == nil {
				nets = append(nets, ipNet)
			}
			continue
		}
		if ip := net.ParseIP(p); ip != nil {
			ips = append(ips, ip)
		}
	}

	return &ProxyChecker{ips: ips, nets: nets}
}

// IsTrusted проверяет, входит ли IP в список доверенных прокси.
func (c *ProxyChecker) IsTrusted(ip string) bool {
	if ip == "" {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, trusted := range c.ips {
		if trusted.Equal(parsed) {
			return true
		}
	}
	for _, net := range c.nets {
		if net.Contains(parsed) {
			return true
		}
	}
	return false
}

// ClientIP определяет реальный IP клиента по адресу соединения и заголовкам.
//
// Если соединение пришло НЕ от доверенного прокси, заголовки
// X-Forwarded-For / X-Real-IP не учитываются: их может подделать клиент.
//
// Если соединение пришло от доверенного прокси, IP берётся из XFF методом
// rightmost-untrusted: цепочка обходится справа налево, пропуская доверенные
// прокси; возвращается первый недоверенный адрес. Это защищает от подмены:
// клиент может дописать свои адреса в начало XFF, но Caddy дописывает
// реальный IP клиента последним, поэтому последний недоверенный адрес —
// настоящий клиент.
func ClientIP(checker *ProxyChecker, header http.Header, remoteAddr string) string {
	peer := IPFromAddr(remoteAddr)

	if checker == nil || !checker.IsTrusted(peer) {
		return peer
	}

	if xff := header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(parts[i])
			if ip == "" {
				continue
			}
			if !checker.IsTrusted(ip) {
				return ip
			}
		}
		// Вся цепочка состоит из доверенных прокси — возвращаем последний элемент,
		// это адрес, с которого прокси получил запрос.
		return strings.TrimSpace(parts[len(parts)-1])
	}

	if xri := strings.TrimSpace(header.Get("X-Real-IP")); xri != "" {
		return xri
	}

	return peer
}

// IPFromAddr извлекает IP из адреса, который может содержать порт.
// Примеры: "192.168.1.1:12345" -> "192.168.1.1", "[::1]:8080" -> "::1"
func IPFromAddr(addr string) string {
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
