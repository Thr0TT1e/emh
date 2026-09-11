package netutil

import (
	"net/http"
	"testing"
)

func makeHeader(kv map[string]string) http.Header {
	h := http.Header{}
	for k, v := range kv {
		h.Set(k, v)
	}
	return h
}

// TestClientIP_UntrustedPeerIgnoresXFF: XFF от недоверенного адреса игнорируется.
func TestClientIP_UntrustedPeerIgnoresXFF(t *testing.T) {
	checker := NewProxyChecker([]string{"10.0.0.5"})
	got := ClientIP(checker, makeHeader(map[string]string{
		"X-Forwarded-For": "1.2.3.4",
		"X-Real-IP":       "5.6.7.8",
	}), "198.51.100.1:4444")
	if got != "198.51.100.1" {
		t.Errorf("got %q, want peer IP 198.51.100.1", got)
	}
}

// TestClientIP_TrustedProxyRightmostUntrusted: клиентский спуфинг слева от XFF
// не работает — берётся последний недоверенный адрес.
func TestClientIP_TrustedProxyRightmostUntrusted(t *testing.T) {
	checker := NewProxyChecker([]string{"10.0.0.5"})
	got := ClientIP(checker, makeHeader(map[string]string{
		"X-Forwarded-For": "6.6.6.6, 203.0.113.42, 10.0.0.5",
	}), "10.0.0.5:443")
	if got != "203.0.113.42" {
		t.Errorf("got %q, want 203.0.113.42", got)
	}
}

// TestClientIP_TrustedProxySingleXFF: одиночный XFF от доверенного прокси.
func TestClientIP_TrustedProxySingleXFF(t *testing.T) {
	checker := NewProxyChecker([]string{"10.0.0.0/8"})
	got := ClientIP(checker, makeHeader(map[string]string{
		"X-Forwarded-For": "203.0.113.42",
	}), "10.1.2.3:443")
	if got != "203.0.113.42" {
		t.Errorf("got %q, want 203.0.113.42", got)
	}
}

// TestClientIP_AllTrustedChain: вся цепочка доверенная — возвращается
// последний элемент (адрес, с которым прокси получил запрос).
func TestClientIP_AllTrustedChain(t *testing.T) {
	checker := NewProxyChecker([]string{"10.0.0.0/8"})
	got := ClientIP(checker, makeHeader(map[string]string{
		"X-Forwarded-For": "10.0.0.5, 10.0.0.6",
	}), "10.0.0.6:443")
	if got != "10.0.0.6" {
		t.Errorf("got %q, want 10.0.0.6 (последний элемент цепочки)", got)
	}
}
func TestClientIP_XRealIPFallback(t *testing.T) {
	checker := NewProxyChecker([]string{"10.0.0.5"})
	got := ClientIP(checker, makeHeader(map[string]string{
		"X-Real-IP": "203.0.113.42",
	}), "10.0.0.5:443")
	if got != "203.0.113.42" {
		t.Errorf("got %q, want 203.0.113.42", got)
	}
}

// TestProxyChecker точные IP и CIDR.
func TestProxyChecker(t *testing.T) {
	checker := NewProxyChecker([]string{"127.0.0.1", "::1", "172.16.0.0/12", "  ", "not-an-ip"})

	cases := map[string]bool{
		"127.0.0.1": true,
		"::1":       true,
		"172.17.0.9": true,
		"172.32.0.1": false,
		"10.0.0.1":   false,
		"":           false,
		"garbage":    false,
	}
	for ip, want := range cases {
		if got := checker.IsTrusted(ip); got != want {
			t.Errorf("IsTrusted(%q) = %v, want %v", ip, got, want)
		}
	}
}

// TestIPFromAddr извлечение IP из адреса с портом и без.
func TestIPFromAddr(t *testing.T) {
	cases := map[string]string{
		"192.168.1.1:12345": "192.168.1.1",
		"[::1]:8080":        "::1",
		"192.168.1.1":       "192.168.1.1",
		"":                  "",
	}
	for addr, want := range cases {
		if got := IPFromAddr(addr); got != want {
			t.Errorf("IPFromAddr(%q) = %q, want %q", addr, got, want)
		}
	}
}
