package eav

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Blackduck struct {
	Namespace string
	KBAddress string
}

func (bd *Blackduck) DenyAll() *Policy {
	// Explanation: blanket deny of *everything* to and from Blackduck
	return &Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deny-all-blackduck-traffic",
		},
		Spec: PolicySpec{
			Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
			Priority:      0,
			TrafficMatcher: &TrafficEdge{
				Type: TrafficMatchTypeAny,
				Source: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: bd.Namespace},
					},
				},
				Dest: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: bd.Namespace},
					},
				},
			},
			Directive: DirectiveDeny,
		},
	}
}

func (bd *Blackduck) AllowDNSOnTCP() *Policy {
	// Explanation: allow DNS from the Blackduck namespace
	return &Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allow-dns-from-blackduck",
		},
		Spec: PolicySpec{
			Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Priority:      10,
			TrafficMatcher: &TrafficEdge{
				Type: TrafficMatchTypeAll,
				Source: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: bd.Namespace},
					},
				},
				Port:     NumberedPortMatcher(53),
				Protocol: &ProtocolMatcher{Values: []v1.Protocol{v1.ProtocolTCP}},
			},
			Directive: DirectiveAllow,
		},
	}
}

func (bd *Blackduck) AllowEgressToKB() *Policy {
	// Explanation: if source is the Blackduck namespace, and the destination
	// is the KB API, allow the traffic
	return &Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allow-egress-to-kb",
		},
		Spec: PolicySpec{
			Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Priority:      10,
			TrafficMatcher: &TrafficEdge{
				Type: TrafficMatchTypeAll,
				Source: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: bd.Namespace},
					},
				},
				Dest: &PeerMatcher{
					IP: &IPMatcher{
						Value: &bd.KBAddress,
					},
				},
			},
			Directive: DirectiveAllow,
		},
	}
}

func (bd *Blackduck) AllowBDNamespaceCommunication() *Policy {
	// Explanation: if both source and destination are the Blackduck namespace,
	// allow the traffic
	return &Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "all-blackduck-internamespace-communication",
		},
		Spec: PolicySpec{
			Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
			Priority:      10,
			TrafficMatcher: &TrafficEdge{
				Type: TrafficMatchTypeAll,
				Source: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: bd.Namespace},
					},
				},
				Dest: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: bd.Namespace},
					},
				},
			},
			Directive: DirectiveAllow,
		},
	}
}
