package analyzer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

var blockedCIDRs = mustPrefixes([]string{
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"169.254.0.0/16",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
})

type ValidatedURL struct {
	Source     string
	Normalized string
	Hostname   string
	Scheme     string
	Parsed     *url.URL
}

func ValidateURL(ctx context.Context, raw string) (ValidatedURL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ValidatedURL{}, errors.New("url is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return ValidatedURL{}, fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme == "" {
		return ValidatedURL{}, errors.New("url must include http or https scheme")
	}
	if parsed.User != nil {
		return ValidatedURL{}, errors.New("urls with embedded credentials are not allowed")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ValidatedURL{}, fmt.Errorf("unsupported url scheme %q", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return ValidatedURL{}, errors.New("url hostname is required")
	}
	if err := validateHost(ctx, parsed.Hostname()); err != nil {
		return ValidatedURL{}, err
	}
	normalized := *parsed
	if normalized.Path == "" {
		normalized.Path = "/"
	}
	return ValidatedURL{
		Source: raw, Normalized: normalized.String(), Hostname: parsed.Hostname(), Scheme: parsed.Scheme, Parsed: &normalized,
	}, nil
}

func validateHost(ctx context.Context, host string) error {
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return errors.New("localhost targets are blocked by default")
	}
	if ip, err := netip.ParseAddr(lower); err == nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("private or reserved ip %s is blocked", ip.String())
		}
		return nil
	}
	resolver := net.Resolver{}
	addrs, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("resolve host %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("no public ip addresses resolved for %q", host)
	}
	for _, addr := range addrs {
		ip, ok := netip.AddrFromSlice(addr)
		if !ok {
			continue
		}
		if isBlockedIP(ip) {
			return fmt.Errorf("resolved private or reserved ip %s for %q", ip.String(), host)
		}
	}
	return nil
}

func isBlockedIP(ip netip.Addr) bool {
	for _, prefix := range blockedCIDRs {
		if prefix.Contains(ip) {
			return true
		}
	}
	return ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsPrivate() || ip.IsMulticast() || ip.IsUnspecified()
}

func mustPrefixes(values []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			panic(err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}
