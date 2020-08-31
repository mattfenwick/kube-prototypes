package eav

import networkingv1 "k8s.io/api/networking/v1"

type Policy struct {
	Type          networkingv1.PolicyType
	TargetMatcher PeerMatcher
	PeerMatcher   PeerMatcher
	PortMatcher   *ProtocolPortMatcher
}

func (p *Policy) IsMatchForTarget(target *Peer) bool {
	return p.TargetMatcher.IsPeerMatch(target)
}

// Allows returns:
// - `false, false` if the policy doesn't match the traffic target
// - `true, false` if the policy matches the traffic target, but doesn't allow the traffic
// - `true, true` if the policy matches the traffic target, and *does* allow the traffic
// In order to allow the traffic, ALL of the following must be matched:
// - Peer
// - Port
// - Protocol
func (p *Policy) Allows(t *Traffic) (bool, bool) {
	var isPeerMatch bool
	switch p.Type {
	case networkingv1.PolicyTypeIngress:
		if !p.IsMatchForTarget(t.Destination) {
			return false, false
		}
		isPeerMatch = p.PeerMatcher.IsPeerMatch(t.Source)
	case networkingv1.PolicyTypeEgress:
		if !p.IsMatchForTarget(t.Source) {
			return false, false
		}
		isPeerMatch = p.PeerMatcher.IsPeerMatch(t.Destination)
	default:
		panic("invalid policy type")
	}
	return true, isPeerMatch && p.PortMatcher.IsProtocolPortMatch(t.Protocol, t.Port)
}
