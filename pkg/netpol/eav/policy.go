package eav

import networkingv1 "k8s.io/api/networking/v1"

type Policy struct {
	Type          networkingv1.PolicyType
	TargetMatcher PeerMatcher
	PeerMatcher   PeerMatcher
	PortMatcher   ProtoportMatcher
}

func (p *Policy) IsMatchForTarget(target *Peer) bool {
	return p.TargetMatcher.IsPeerMatch(target)
}

// Allows returns:
// - `false, false` if the policy doesn't match the traffic target
// - `true, false` if the policy matches the traffic and doesn't allow it
// - `true, true` if the policy matches the traffic and *does* allow it
func (p *Policy) Allows(t *Traffic) (bool, bool) {
	switch p.Type {
	case networkingv1.PolicyTypeIngress:
		if !p.IsMatchForTarget(t.Destination) {
			return false, false
		}
		return true, p.PeerMatcher.IsPeerMatch(t.Source) && p.PortMatcher.IsProtocolPortMatch(t.Protocol, t.Port)
	case networkingv1.PolicyTypeEgress:
		if !p.IsMatchForTarget(t.Source) {
			return false, false
		}
		return true, p.PeerMatcher.IsPeerMatch(t.Destination) && p.PortMatcher.IsProtocolPortMatch(t.Protocol, t.Port)
	default:
		panic("invalid policy type")
	}
}
