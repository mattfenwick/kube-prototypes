package eav

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
			matcher, directive := BuildTrafficPeersFromIngress(netpol.Namespace, netpol.Spec.Ingress)
			policies = append(policies, &Policy{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: PolicySpec{
					Compatibility:  []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
					TrafficMatcher: matcher,
					Directive:      directive,
				},
			})
		case networkingv1.PolicyTypeEgress:
			matcher, directive := BuildTrafficPeersFromEgress(netpol.Namespace, netpol.Spec.Egress)
			policies = append(policies, &Policy{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: PolicySpec{
					Compatibility:  []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
					TrafficMatcher: matcher,
					Directive:      directive,
				},
			})
		}
	}
	return policies
}

func BuildTrafficPeersFromIngress(policyNamespace string, ingresses []networkingv1.NetworkPolicyIngressRule) (TrafficMatcher, Directive) {
	if len(ingresses) == 0 {
		return EverythingMatcher, DirectiveDeny
	}

	var sdaps []TrafficMatcher
	for _, ingress := range ingresses {
		sdaps = append(sdaps, BuildSourceDestAndPorts(SourceSelector, policyNamespace, ingress.Ports, ingress.From))
	}
	return NewAny(sdaps...), DirectiveAllow
}

func BuildTrafficPeersFromEgress(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) (TrafficMatcher, Directive) {
	if len(egresses) == 0 {
		// This is a hard-to-grok case:
		//   in network policy land, it should select NO Peers, with the effect of blocking all communication to target
		//   in matcher land, we translate that by matching ALL with a deny
		return EverythingMatcher, DirectiveDeny
	}

	var sdaps []TrafficMatcher
	for _, egress := range egresses {
		sdaps = append(sdaps, BuildSourceDestAndPorts(DestSelector, policyNamespace, egress.Ports, egress.To))
	}
	return NewAny(sdaps...), DirectiveAllow
}

func BuildSourceDestAndPorts(selector string, policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) TrafficMatcher {
	sds := BuildSourceDestsFromSlice(selector, policyNamespace, peers)
	if len(npPorts) == 0 {
		return sds
	}
	ports := BuildPortsFromSlice(npPorts)
	return NewAll(sds, ports)
}

func BuildPortsFromSlice(npPorts []networkingv1.NetworkPolicyPort) TrafficMatcher {
	if len(npPorts) == 0 {
		panic("can't handle 0 NetworkPolicyPorts")
	}
	var terms []TrafficMatcher
	for _, p := range npPorts {
		terms = append(terms, BuildPort(p))
	}
	return NewAny(terms...)
}

func BuildSourceDestsFromSlice(selector string, policyNamespace string, peers []networkingv1.NetworkPolicyPeer) TrafficMatcher {
	if len(peers) == 0 {
		return EverythingMatcher
	}
	var terms []TrafficMatcher
	for _, from := range peers {
		terms = append(terms, BuildSourceDest(selector, policyNamespace, from))
	}
	return NewAny(terms...)
}

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

func BuildSourceDest(path string, policyNamespace string, peer networkingv1.NetworkPolicyPeer) TrafficMatcher {
	if peer.IPBlock != nil {
		return IPBlockMatcher(NewKeyPathSelector(path, IPSelector), peer.IPBlock.CIDR, peer.IPBlock.Except)
	}
	podSel := peer.PodSelector
	nsSel := peer.NamespaceSelector
	isInternalMatcher := &Not{isExternalMatcher(path)}
	policyNamespaceMatcher := NewEqual(NewKeyPathSelector(path, InternalSelector, NamespaceSelector), &ConstantSelector{policyNamespace})
	if podSel == nil || isLabelSelectorEmpty(*podSel) {
		if nsSel == nil {
			return NewAll(isInternalMatcher, policyNamespaceMatcher)
		} else if isLabelSelectorEmpty(*nsSel) {
			return isInternalMatcher
		} else {
			// nsSel has some stuff
			return NewAll(
				isInternalMatcher,
				KubeMatchLabelSelector(NewKeyPathSelector(path, InternalSelector, NamespaceLabelsSelector), *nsSel))
		}
	} else {
		podLabelsMatcher := KubeMatchLabelSelector(NewKeyPathSelector(path, InternalSelector, PodLabelsSelector), *podSel)
		// podSel has some stuff
		if nsSel == nil {
			return NewAll(
				isInternalMatcher,
				policyNamespaceMatcher,
				podLabelsMatcher)
		} else if isLabelSelectorEmpty(*nsSel) {
			return NewAll(
				isInternalMatcher,
				KubeMatchLabelSelector(NewKeyPathSelector(path, InternalSelector, PodLabelsSelector), *podSel))
		} else {
			// nsSel has some stuff
			return NewAll(
				isInternalMatcher,
				KubeMatchLabelSelector(NewKeyPathSelector(path, InternalSelector, PodLabelsSelector), *podSel),
				KubeMatchLabelSelector(NewKeyPathSelector(path, InternalSelector, NamespaceLabelsSelector), *nsSel))
		}
	}
}

func BuildPort(p networkingv1.NetworkPolicyPort) TrafficMatcher {
	protocol := v1.ProtocolTCP
	if p.Protocol != nil {
		protocol = *p.Protocol
	}
	if p.Port == nil {
		return ProtocolMatcher(protocol)
	}
	var portMatcher TrafficMatcher
	switch p.Port.Type {
	case intstr.Int:
		portMatcher = NumberedPortMatcher(int(p.Port.IntVal))
	case intstr.String:
		portMatcher = NamedPortMatcher(p.Port.StrVal)
	default:
		panic("invalid intstr type")
	}
	return NewAll(ProtocolMatcher(protocol), portMatcher)
}
