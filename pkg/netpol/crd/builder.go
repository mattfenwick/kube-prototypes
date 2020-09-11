package crd

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildNetworkPolicy(policy *networkingv1.NetworkPolicy) *Policies {
	return BuildPolicies([]*networkingv1.NetworkPolicy{policy})
}

func BuildPolicies(netpols []*networkingv1.NetworkPolicy) *Policies {
	var policies []*Policy
	for _, policy := range netpols {
		policies = append(policies, BuildTarget(policy)...)
	}
	return &Policies{Policies: policies}
}

func BuildTarget(netpol *networkingv1.NetworkPolicy) []*Policy {
	var policies []*Policy
	for _, pType := range netpol.Spec.PolicyTypes {
		switch pType {
		case networkingv1.PolicyTypeIngress:
			edges, directive := BuildTrafficPeersFromIngress(netpol)
			for _, edge := range edges {
				policies = append(policies, &Policy{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: PolicySpec{
						Compatibility:  []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
						TrafficMatcher: edge,
						Directive:      directive,
						Priority:       0,
					},
				})
			}
		case networkingv1.PolicyTypeEgress:
			edges, directive := BuildTrafficPeersFromEgress(netpol)
			for _, edge := range edges {
				policies = append(policies, &Policy{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: PolicySpec{
						Compatibility:  []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
						TrafficMatcher: edge,
						Directive:      directive,
						Priority:       0,
					},
				})
			}
		}
	}
	return policies
}

func BuildTrafficPeersFromIngress(netpol *networkingv1.NetworkPolicy) ([]*TrafficEdge, Directive) {
	if len(netpol.Spec.Ingress) == 0 {
		return []*TrafficEdge{EverythingMatcher}, DirectiveDeny
	}

	var edges []*TrafficEdge
	for _, ingress := range netpol.Spec.Ingress {
		edges = append(edges, BuildSourceDestAndPorts(true, netpol.Spec.PodSelector, netpol.Namespace, ingress.Ports, ingress.From)...)
	}
	return edges, DirectiveAllow
}

func BuildTrafficPeersFromEgress(netpol *networkingv1.NetworkPolicy) ([]*TrafficEdge, Directive) {
	if len(netpol.Spec.Egress) == 0 {
		// This is a hard-to-grok case:
		//   in network policy land, it should select NO Peers, with the effect of blocking all communication to target
		//   in matcher land, we translate that by matching ALL with a deny
		return []*TrafficEdge{EverythingMatcher}, DirectiveDeny
	}

	var edges []*TrafficEdge
	for _, egress := range netpol.Spec.Egress {
		edges = append(edges, BuildSourceDestAndPorts(false, netpol.Spec.PodSelector, netpol.Namespace, egress.Ports, egress.To)...)
	}
	return edges, DirectiveAllow
}

func BuildSourceDestAndPorts(isIngress bool, targetPodSelector metav1.LabelSelector, policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) []*TrafficEdge {
	var ports []*PortMatcher
	var protocols []*ProtocolMatcher
	sds := BuildSourceDestsFromSlice(policyNamespace, peers)
	if len(npPorts) > 0 {
		ports, protocols = BuildPortsFromSlice(npPorts)
	}

	// handle a couple of corner cases
	if len(ports) == 0 {
		ports = []*PortMatcher{nil}
	}
	if len(protocols) == 0 {
		protocols = []*ProtocolMatcher{nil}
	}

	var edges []*TrafficEdge
	for _, peer := range sds {
		for _, port := range ports {
			for _, protocol := range protocols {
				target := &PeerMatcher{
					RelativeLocation: &PeerLocationInternal,
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: policyNamespace},
						PodLabels: &targetPodSelector,
					},
				}
				var source, dest *PeerMatcher
				if isIngress {
					source = peer
					dest = target
				} else {
					// egress
					source = target
					dest = peer
				}
				edges = append(edges, &TrafficEdge{
					Type:     TrafficMatchTypeAll,
					Source:   source,
					Dest:     dest,
					Port:     port,
					Protocol: protocol,
				})
			}
		}
	}

	return edges
}

func BuildPortsFromSlice(npPorts []networkingv1.NetworkPolicyPort) ([]*PortMatcher, []*ProtocolMatcher) {
	if len(npPorts) == 0 {
		panic("can't handle 0 NetworkPolicyPorts")
	}
	var protocols []*ProtocolMatcher
	var ports []*PortMatcher
	for _, p := range npPorts {
		protocalMatcher, portMatcher := BuildPort(p)
		if portMatcher != nil {
			ports = append(ports, portMatcher)
		}
		protocols = append(protocols, protocalMatcher)
	}
	return ports, protocols
}

func BuildSourceDestsFromSlice(policyNamespace string, peers []networkingv1.NetworkPolicyPeer) []*PeerMatcher {
	var terms []*PeerMatcher
	if len(peers) == 0 {
		return append(terms, &PeerMatcher{})
	}
	for _, from := range peers {
		terms = append(terms, BuildSourceDest(policyNamespace, from))
	}
	return terms
}

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

func BuildSourceDest(policyNamespace string, peer networkingv1.NetworkPolicyPeer) *PeerMatcher {
	if peer.IPBlock != nil {
		return &PeerMatcher{IP: &IPMatcher{Block: peer.IPBlock}}
	}
	podSel := peer.PodSelector
	nsSel := peer.NamespaceSelector
	if podSel == nil || isLabelSelectorEmpty(*podSel) {
		if nsSel == nil {
			return &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
				Internal: &InternalPeerMatcher{
					Namespace: &StringMatcher{Value: policyNamespace},
				},
			}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
			}
		} else {
			// nsSel has some stuff
			return &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
				Internal: &InternalPeerMatcher{
					NamespaceLabels: nsSel,
				},
			}
		}
	} else {
		// podSel has some stuff
		if nsSel == nil {
			return &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
				Internal: &InternalPeerMatcher{
					Namespace: &StringMatcher{Value: policyNamespace},
					PodLabels: podSel,
				},
			}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
				Internal: &InternalPeerMatcher{
					PodLabels: podSel,
				},
			}
		} else {
			// nsSel has some stuff
			return &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
				Internal: &InternalPeerMatcher{
					NamespaceLabels: nsSel,
					PodLabels:       podSel,
				},
			}
		}
	}
}

func BuildPort(p networkingv1.NetworkPolicyPort) (*ProtocolMatcher, *PortMatcher) {
	var protocolMatcher *ProtocolMatcher
	var portMatcher *PortMatcher
	protocol := v1.ProtocolTCP
	if p.Protocol != nil {
		protocol = *p.Protocol
	}
	protocolMatcher = &ProtocolMatcher{Values: []v1.Protocol{protocol}}
	if p.Port == nil {
		return protocolMatcher, nil
	}
	switch p.Port.Type {
	case intstr.Int:
		portMatcher = NumberedPortMatcher(int(p.Port.IntVal))
	case intstr.String:
		portMatcher = NamedPortMatcher(p.Port.StrVal)
	default:
		panic("invalid intstr type")
	}
	return protocolMatcher, portMatcher
}
