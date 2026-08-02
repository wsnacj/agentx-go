package networkpolicy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// ErrOutboundPolicyBlocked identifies destinations denied by an outbound policy.
var ErrOutboundPolicyBlocked = errors.New("outbound policy blocked destination")

type OutboundOptions struct {
	AllowPrivateHosts bool
	AllowCIDRs        []string
	DenyCIDRs         []string
	AllowPorts        []int
	DenyPorts         []int
	DefaultDenyCIDRs  []string
	DefaultDenyPorts  []int
	ResolveHostIPs    func(context.Context, string) ([]net.IP, error)
}

type OutboundPolicy struct {
	AllowPrivate     bool
	AllowCIDRs       []*net.IPNet
	AllowCIDRsSet    bool
	DenyCIDRs        []*net.IPNet
	DefaultDenyCIDRs []*net.IPNet
	AllowPorts       map[int]bool
	AllowPortsSet    bool
	DenyPorts        map[int]bool
	ResolveHostIPs   func(context.Context, string) ([]net.IP, error)
}

var (
	DefaultOutboundDeniedPorts = []int{22, 23, 25, 3306, 5432, 6379, 11211, 27017}
	DefaultOutboundDeniedCIDRs = []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"240.0.0.0/4",
		"::1/128",
		"::/128",
		"fc00::/7",
		"fe80::/10",
		"ff00::/8",
		"2001:db8::/32",
	}
)

func NewOutboundPolicy(opts OutboundOptions) (OutboundPolicy, error) {
	allowCIDRs, err := ParseCIDRs(opts.AllowCIDRs)
	if err != nil {
		return OutboundPolicy{}, fmt.Errorf("parse allow cidrs: %w", err)
	}
	denyCIDRs, err := ParseCIDRs(opts.DenyCIDRs)
	if err != nil {
		return OutboundPolicy{}, fmt.Errorf("parse deny cidrs: %w", err)
	}
	defaultDenyCIDRs, err := ParseCIDRs(opts.DefaultDenyCIDRs)
	if err != nil {
		return OutboundPolicy{}, fmt.Errorf("parse default deny cidrs: %w", err)
	}
	denyPorts := BuildPortSet(opts.DenyPorts)
	if len(denyPorts) == 0 {
		denyPorts = BuildPortSet(opts.DefaultDenyPorts)
	}
	return OutboundPolicy{
		AllowPrivate:     opts.AllowPrivateHosts,
		AllowCIDRs:       allowCIDRs,
		AllowCIDRsSet:    opts.AllowCIDRs != nil,
		DenyCIDRs:        denyCIDRs,
		DefaultDenyCIDRs: defaultDenyCIDRs,
		AllowPorts:       BuildPortSet(opts.AllowPorts),
		AllowPortsSet:    opts.AllowPorts != nil,
		DenyPorts:        denyPorts,
		ResolveHostIPs:   opts.ResolveHostIPs,
	}, nil
}

func (p OutboundPolicy) ValidateURL(ctx context.Context, raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if parsed == nil || parsed.Host == "" {
		return nil, fmt.Errorf("invalid url")
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("only http/https is supported")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("url with userinfo is not allowed")
	}
	hostname := strings.TrimSpace(parsed.Hostname())
	if hostname == "" {
		return nil, fmt.Errorf("missing hostname")
	}
	port, err := ResolveURLPort(parsed)
	if err != nil {
		return nil, err
	}
	if err := p.ValidateHostPort(ctx, hostname, port, p.AllowPrivate); err != nil {
		return nil, err
	}
	return parsed, nil
}

func (p OutboundPolicy) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return p.DialContextWithAllowPrivate(ctx, network, address, p.AllowPrivate)
}

func (p OutboundPolicy) DialContextWithAllowPrivate(ctx context.Context, networkName string, address string, allowPrivate bool) (net.Conn, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	if err := p.ValidatePort(port); err != nil {
		return nil, err
	}
	ips, err := p.ResolveIPs(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("resolve hostname: no ip")
	}
	dialer := &net.Dialer{}
	var lastErr error
	for _, ip := range ips {
		if err := p.ValidateIPWithAllowPrivate(ip, allowPrivate); err != nil {
			lastErr = err
			continue
		}
		conn, dialErr := dialer.DialContext(ctx, networkName, net.JoinHostPort(ip.String(), portText))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no allowed address")
}

func (p OutboundPolicy) ValidateHostPort(ctx context.Context, hostname string, port int, allowPrivate bool) error {
	if strings.TrimSpace(hostname) == "" {
		return fmt.Errorf("missing hostname")
	}
	if err := p.ValidatePort(port); err != nil {
		return err
	}
	ips, err := p.ResolveIPs(ctx, hostname)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return fmt.Errorf("resolve hostname: no ip")
	}
	for _, ip := range ips {
		if err := p.ValidateIPWithAllowPrivate(ip, allowPrivate); err != nil {
			return err
		}
	}
	return nil
}

func (p OutboundPolicy) ResolveIPs(ctx context.Context, hostname string) ([]net.IP, error) {
	name := strings.TrimSpace(hostname)
	if name == "" {
		return nil, fmt.Errorf("missing hostname")
	}
	if ip := net.ParseIP(name); ip != nil {
		return []net.IP{ip}, nil
	}
	resolved := []net.IP{}
	if p.ResolveHostIPs != nil {
		ips, err := p.ResolveHostIPs(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("resolve hostname: %w", err)
		}
		resolved = append(resolved, ips...)
	} else {
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("resolve hostname: %w", err)
		}
		for _, ipAddr := range ips {
			resolved = append(resolved, ipAddr.IP)
		}
	}
	unique := map[string]net.IP{}
	for _, ip := range resolved {
		if ip == nil {
			continue
		}
		key := ip.String()
		if _, ok := unique[key]; ok {
			continue
		}
		unique[key] = append(net.IP(nil), ip...)
	}
	out := make([]net.IP, 0, len(unique))
	for _, ip := range unique {
		out = append(out, ip)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out, nil
}

func (p OutboundPolicy) ValidateIP(ip net.IP) error {
	return p.ValidateIPWithAllowPrivate(ip, p.AllowPrivate)
}

func (p OutboundPolicy) ValidateIPWithAllowPrivate(ip net.IP, allowPrivate bool) error {
	if ip == nil {
		return fmt.Errorf("invalid ip")
	}
	if !allowPrivate && IsPrivateOrLocalIP(ip) {
		return fmt.Errorf("%w: private address is not allowed", ErrOutboundPolicyBlocked)
	}
	for _, cidr := range p.DenyCIDRs {
		if cidr.Contains(ip) {
			return fmt.Errorf("%w: address is blocked by cidr policy", ErrOutboundPolicyBlocked)
		}
	}
	if !allowPrivate {
		for _, cidr := range p.DefaultDenyCIDRs {
			if cidr.Contains(ip) {
				return fmt.Errorf("%w: address is blocked by cidr policy", ErrOutboundPolicyBlocked)
			}
		}
	}
	if p.AllowCIDRsSet {
		allowed := false
		for _, cidr := range p.AllowCIDRs {
			if cidr.Contains(ip) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: address not allowed by cidr policy", ErrOutboundPolicyBlocked)
		}
	}
	return nil
}

func (p OutboundPolicy) ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port")
	}
	if p.AllowPortsSet && !p.AllowPorts[port] {
		return fmt.Errorf("%w: port is not allowed", ErrOutboundPolicyBlocked)
	}
	if p.DenyPorts[port] {
		return fmt.Errorf("%w: port is blocked", ErrOutboundPolicyBlocked)
	}
	return nil
}

func BuildPortSet(values []int) map[int]bool {
	if len(values) == 0 {
		return nil
	}
	set := map[int]bool{}
	for _, value := range values {
		if value <= 0 || value > 65535 {
			continue
		}
		set[value] = true
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func ParseCIDRs(values []string) ([]*net.IPNet, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		_, network, err := net.ParseCIDR(trimmed)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", trimmed, err)
		}
		out = append(out, network)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func ResolveURLPort(parsed *url.URL) (int, error) {
	return ResolveURLPortWithDefaults(parsed, map[string]int{
		"http":  80,
		"https": 443,
	})
}

func ResolveURLPortWithDefaults(parsed *url.URL, defaults map[string]int) (int, error) {
	portText := strings.TrimSpace(parsed.Port())
	if portText != "" {
		port, err := strconv.Atoi(portText)
		if err != nil {
			return 0, fmt.Errorf("invalid port")
		}
		return port, nil
	}
	if parsed == nil {
		return 0, fmt.Errorf("unsupported scheme")
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if defaults != nil {
		if port, ok := defaults[scheme]; ok {
			return port, nil
		}
	}
	return 0, fmt.Errorf("unsupported scheme")
}

func IsPrivateOrLocalIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	return ip.IsInterfaceLocalMulticast()
}
