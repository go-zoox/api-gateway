package ippolicy

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func parsePrefixList(cidrList []string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, s := range cidrList {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		p, err := parseOneCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", s, err)
		}
		out = append(out, p)
	}
	return out, nil
}

func parseOneCIDR(s string) (netip.Prefix, error) {
	if strings.Contains(s, "/") {
		return netip.ParsePrefix(s)
	}
	a, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Prefix{}, err
	}
	bits := 32
	if a.Is6() {
		bits = 128
	}
	return a.Prefix(bits)
}

// DirectPeerIP returns the address of the immediate TCP client (the peer connected to the gateway).
func DirectPeerIP(req *http.Request) (netip.Addr, error) {
	if req == nil {
		return netip.Addr{}, fmt.Errorf("nil request")
	}
	if req.RemoteAddr == "" {
		return netip.Addr{}, fmt.Errorf("empty remote addr")
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return netip.ParseAddr(strings.TrimSpace(req.RemoteAddr))
	}
	return netip.ParseAddr(host)
}

// ClientIP returns the end-client address. When direct is in trusted, X-Forwarded-For (leftmost) or
// X-Real-IP is used; otherwise direct is returned.
func ClientIP(req *http.Request, direct netip.Addr, trusted []netip.Prefix) netip.Addr {
	if !addrInAnyPrefix(direct, trusted) {
		return direct
	}
	xff := req.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			first := strings.TrimSpace(parts[0])
			if a, err := netip.ParseAddr(first); err == nil {
				return a
			}
		}
	}
	if rip := strings.TrimSpace(req.Header.Get("X-Real-IP")); rip != "" {
		if a, err := netip.ParseAddr(rip); err == nil {
			return a
		}
	}
	return direct
}

func addrInAnyPrefix(a netip.Addr, prefixes []netip.Prefix) bool {
	if !a.IsValid() {
		return false
	}
	if len(prefixes) == 0 {
		return false
	}
	for _, p := range prefixes {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

type compiled struct {
	allow            []netip.Prefix
	deny             []netip.Prefix
	trusted          []netip.Prefix
	message          string
	allowListNonEmpty bool
}

func (c *compiled) allows(ip netip.Addr) bool {
	if !ip.IsValid() {
		return false
	}
	for _, d := range c.deny {
		if d.Contains(ip) {
			return false
		}
	}
	if !c.allowListNonEmpty {
		return true
	}
	for _, a := range c.allow {
		if a.Contains(ip) {
			return true
		}
	}
	return false
}

func newCompiled(allow, deny, trusted []string, message string) (*compiled, error) {
	c := &compiled{message: message}
	var err error
	if c.deny, err = parsePrefixList(deny); err != nil {
		return nil, err
	}
	if c.allow, err = parsePrefixList(allow); err != nil {
		return nil, err
	}
	c.allowListNonEmpty = len(c.allow) > 0
	if c.trusted, err = parsePrefixList(trusted); err != nil {
		return nil, err
	}
	if c.message == "" {
		c.message = "Forbidden"
	}
	return c, nil
}
