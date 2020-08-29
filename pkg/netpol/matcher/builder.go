package matcher

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
		sdaps = append(sdaps, BuildSourceDestAndPortsFromIngressRule(policyNamespace, ingress)...)
	}
	return &TrafficPeers{SourcesOrDests: sdaps}
}

func BuildTrafficPeersFromEgress(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) *TrafficPeers {
	var sdaps []*SourceDestAndPort
	for _, egress := range egresses {
		sdaps = append(sdaps, BuildSourceDestAndPortsFromEgressRule(policyNamespace, egress)...)
	}
	return &TrafficPeers{SourcesOrDests: sdaps}
}

func BuildSourceDestAndPortsFromIngressRule(policyNamespace string, ingress networkingv1.NetworkPolicyIngressRule) []*SourceDestAndPort {
	// 1. build ports
	var ports []Port
	if len(ingress.Ports) == 0 {
		ports = append(ports, &AllPortsAllProtocols{})
	} else {
		for _, p := range ingress.Ports {
			ports = append(ports, BuildPort(p))
		}
	}

	// 2. build SourceDests
	var sds []SourceDest
	if len(ingress.From) == 0 {
		sds = append(sds, &AnywhereSourceDest{})
	} else {
		for _, from := range ingress.From {
			sds = append(sds, BuildSourceDest(policyNamespace, from))
		}
	}

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

func BuildSourceDestAndPortsFromEgressRule(policyNamespace string, egress networkingv1.NetworkPolicyEgressRule) []*SourceDestAndPort {
	// 1. build ports
	var ports []Port
	if len(egress.Ports) == 0 {
		ports = append(ports, &AllPortsAllProtocols{})
	} else {
		for _, p := range egress.Ports {
			ports = append(ports, BuildPort(p))
		}
	}

	// 2. build SourceDests
	var sds []SourceDest
	if len(egress.To) == 0 {
		sds = append(sds, &AnywhereSourceDest{})
	} else {
		for _, from := range egress.To {
			sds = append(sds, BuildSourceDest(policyNamespace, from))
		}
	}

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

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

func BuildSourceDest(policyNamespace string, peer networkingv1.NetworkPolicyPeer) SourceDest {
	if peer.IPBlock != nil {
		return &IPBlockSourceDest{
			CIDR:   peer.IPBlock.CIDR,
			Except: peer.IPBlock.Except,
		}
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
			return &AllPodsInNamespaceSourceDest{NamespaceSelector: *nsSel}
		}
	} else {
		// podSel has some stuff
		if nsSel == nil {
			return &PodsInAllNamespacesSourceDest{PodSelector: *podSel}
		} else if isLabelSelectorEmpty(*nsSel) {
			return &PodsInPolicyNamespaceSourceDest{
				PodSelector: *podSel,
				Namespace:   policyNamespace,
			}
		} else {
			// nsSel has some stuff
			return &SpecificPodsNamespaceSourceDest{
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
