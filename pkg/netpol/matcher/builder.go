package matcher

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildNetworkPolicy(policy *networkingv1.NetworkPolicy) *NetworkPolicies {
	return BuildNetworkPolicies([]*networkingv1.NetworkPolicy{policy})
}

func BuildNetworkPolicies(netpols []*networkingv1.NetworkPolicy) *NetworkPolicies {
	np := NewNetworkPolicies()
	for _, policy := range netpols {
		np.AddTarget(BuildTarget(policy))
	}
	return np
}

func BuildTarget(netpol *networkingv1.NetworkPolicy) *Target {
	target := &Target{
		Namespace:   netpol.Namespace,
		PodSelector: netpol.Spec.PodSelector,
	}
	for _, pType := range netpol.Spec.PolicyTypes {
		switch pType {
		case networkingv1.PolicyTypeIngress:
			target.Ingress = BuildTrafficPeersFromIngress(netpol.Namespace, netpol.Spec.Ingress)
		case networkingv1.PolicyTypeEgress:
			target.Egress = BuildTrafficPeersFromEgress(netpol.Namespace, netpol.Spec.Egress)
		}
	}
	return target
}

func BuildTrafficPeersFromIngress(policyNamespace string, ingresses []networkingv1.NetworkPolicyIngressRule) *TrafficPeers {
	var sdaps []*SourceDestAndPort
	for _, ingress := range ingresses {
		sdaps = append(sdaps, BuildSourceDestAndPorts(policyNamespace, ingress.Ports, ingress.From)...)
	}
	return &TrafficPeers{SourcesOrDests: sdaps}
}

func BuildTrafficPeersFromEgress(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) *TrafficPeers {
	var sdaps []*SourceDestAndPort
	for _, egress := range egresses {
		sdaps = append(sdaps, BuildSourceDestAndPorts(policyNamespace, egress.Ports, egress.To)...)
	}
	return &TrafficPeers{SourcesOrDests: sdaps}
}

func BuildSourceDestAndPorts(policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) []*SourceDestAndPort {
	// 1. build ports
	ports := BuildPortsFromSlice(npPorts)
	// 2. build SourceDests
	sds := BuildSourceDestsFromSlice(policyNamespace, peers)
	// 3. build the cartesian product of ports and SourceDests
	var sdaps []*SourceDestAndPort
	for _, port := range ports {
		for _, sd := range sds {
			sdaps = append(sdaps, &SourceDestAndPort{
				SourceDest: sd,
				Port:       port,
			})
		}
	}
	return sdaps
}

func BuildPortsFromSlice(npPorts []networkingv1.NetworkPolicyPort) []Port {
	var ports []Port
	if len(npPorts) == 0 {
		ports = append(ports, &AllPortsAllProtocols{})
	} else {
		for _, p := range npPorts {
			ports = append(ports, BuildPort(p))
		}
	}
	return ports
}

func BuildSourceDestsFromSlice(policyNamespace string, peers []networkingv1.NetworkPolicyPeer) []SourceDest {
	var sds []SourceDest
	if len(peers) == 0 {
		sds = append(sds, &AnywhereSourceDest{})
	} else {
		for _, from := range peers {
			sds = append(sds, BuildSourceDest(policyNamespace, from))
		}
	}
	return sds
}

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

func BuildSourceDest(policyNamespace string, peer networkingv1.NetworkPolicyPeer) SourceDest {
	if peer.IPBlock != nil {
		return &IPBlockSourceDest{peer.IPBlock}
	}
	podSel := peer.PodSelector
	nsSel := peer.NamespaceSelector
	if podSel == nil || isLabelSelectorEmpty(*podSel) {
		if nsSel == nil {
			return &AllPodsInPolicyNamespaceSourceDest{Namespace: policyNamespace}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &AllPodsAllNamespacesSourceDest{}
		} else {
			// nsSel has some stuff
			return &AllPodsInMatchingNamespacesSourceDest{NamespaceSelector: *nsSel}
		}
	} else {
		// podSel has some stuff
		if nsSel == nil {
			return &MatchingPodsInPolicyNamespaceSourceDest{
				PodSelector: *podSel,
				Namespace:   policyNamespace,
			}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &MatchingPodsInAllNamespacesSourceDest{PodSelector: *podSel}
		} else {
			// nsSel has some stuff
			return &MatchingPodsInMatchingNamespacesSourceDest{
				PodSelector:       *podSel,
				NamespaceSelector: *nsSel,
			}
		}
	}
}

func BuildPort(p networkingv1.NetworkPolicyPort) Port {
	protocol := v1.ProtocolTCP
	if p.Protocol != nil {
		protocol = *p.Protocol
	}
	if p.Port == nil {
		return &AllPortsOnProtocol{Protocol: protocol}
	}
	return &PortProtocol{Port: *p.Port, Protocol: protocol}
}
