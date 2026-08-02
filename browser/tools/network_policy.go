package tools

import (
	"context"
	"net"
	"net/url"

	agentxnetwork "github.com/wsnacj/agentx-go/browser/tools/internal/networkpolicy"
)

type outboundNetworkOptions = agentxnetwork.OutboundOptions

type outboundNetworkPolicy struct {
	allowPrivate     bool
	allowCIDRs       []*net.IPNet
	allowCIDRsSet    bool
	denyCIDRs        []*net.IPNet
	defaultDenyCIDRs []*net.IPNet
	allowPorts       map[int]bool
	allowPortsSet    bool
	denyPorts        map[int]bool
}

func newOutboundNetworkPolicy(opts outboundNetworkOptions) (outboundNetworkPolicy, error) {
	shared, err := agentxnetwork.NewOutboundPolicy(agentxnetwork.OutboundOptions(opts))
	if err != nil {
		return outboundNetworkPolicy{}, err
	}
	return outboundNetworkPolicyFromShared(shared), nil
}

func (p outboundNetworkPolicy) validateURL(ctx context.Context, raw string) (*url.URL, error) {
	return p.toShared().ValidateURL(ctx, raw)
}

func (p outboundNetworkPolicy) dialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return p.toShared().DialContext(ctx, network, address)
}

func (p outboundNetworkPolicy) dialContextWithAllowPrivate(ctx context.Context, network string, address string, allowPrivate bool) (net.Conn, error) {
	return p.toShared().DialContextWithAllowPrivate(ctx, network, address, allowPrivate)
}

func (p outboundNetworkPolicy) validateHostPort(ctx context.Context, hostname string, port int, allowPrivate bool) error {
	return p.toShared().ValidateHostPort(ctx, hostname, port, allowPrivate)
}

func (p outboundNetworkPolicy) resolveIPs(ctx context.Context, hostname string) ([]net.IP, error) {
	return p.toShared().ResolveIPs(ctx, hostname)
}

func (p outboundNetworkPolicy) validateIP(ip net.IP) error {
	return p.toShared().ValidateIP(ip)
}

func (p outboundNetworkPolicy) validateIPWithAllowPrivate(ip net.IP, allowPrivate bool) error {
	return p.toShared().ValidateIPWithAllowPrivate(ip, allowPrivate)
}

func (p outboundNetworkPolicy) validatePort(port int) error {
	return p.toShared().ValidatePort(port)
}

func buildPortSet(values []int) map[int]bool {
	return agentxnetwork.BuildPortSet(values)
}

func parseCIDRs(values []string) ([]*net.IPNet, error) {
	return agentxnetwork.ParseCIDRs(values)
}

func resolveURLPort(parsed *url.URL) (int, error) {
	return agentxnetwork.ResolveURLPort(parsed)
}

func resolveURLPortWithDefaults(parsed *url.URL, defaults map[string]int) (int, error) {
	return agentxnetwork.ResolveURLPortWithDefaults(parsed, defaults)
}

var (
	defaultOutboundDeniedPorts = append([]int(nil), agentxnetwork.DefaultOutboundDeniedPorts...)
	defaultOutboundDeniedCIDRs = append([]string(nil), agentxnetwork.DefaultOutboundDeniedCIDRs...)
)

func isPrivateOrLocalIP(ip net.IP) bool {
	return agentxnetwork.IsPrivateOrLocalIP(ip)
}

func (p outboundNetworkPolicy) toShared() agentxnetwork.OutboundPolicy {
	return agentxnetwork.OutboundPolicy{
		AllowPrivate:     p.allowPrivate,
		AllowCIDRs:       p.allowCIDRs,
		AllowCIDRsSet:    p.allowCIDRsSet,
		DenyCIDRs:        p.denyCIDRs,
		DefaultDenyCIDRs: p.defaultDenyCIDRs,
		AllowPorts:       p.allowPorts,
		AllowPortsSet:    p.allowPortsSet,
		DenyPorts:        p.denyPorts,
	}
}

func outboundNetworkPolicyFromShared(policy agentxnetwork.OutboundPolicy) outboundNetworkPolicy {
	return outboundNetworkPolicy{
		allowPrivate:     policy.AllowPrivate,
		allowCIDRs:       policy.AllowCIDRs,
		allowCIDRsSet:    policy.AllowCIDRsSet,
		denyCIDRs:        policy.DenyCIDRs,
		defaultDenyCIDRs: policy.DefaultDenyCIDRs,
		allowPorts:       policy.AllowPorts,
		allowPortsSet:    policy.AllowPortsSet,
		denyPorts:        policy.DenyPorts,
	}
}
