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
				TrafficMatcher: matcher,
				Directive:      directive,
			})
		case networkingv1.PolicyTypeEgress:
			matcher, directive := BuildTrafficPeersFromEgress(netpol.Namespace, netpol.Spec.Egress)
			policies = append(policies, &Policy{
				TrafficMatcher: matcher,
				Directive:      directive,
			})
		}
	}
	return policies
}

func BuildTrafficPeersFromIngress(policyNamespace string, ingresses []networkingv1.NetworkPolicyIngressRule) (TrafficMatcher, Directive) {
	if len(ingresses) == 0 {
		return EverythingMatcher, DirectiveDeny
	}

	family := &SelectorFamily{
		IPSelector:              SourceIPSelector,
		IsExternalSelector:      SourceIsExternalSelector,
		NamespaceMatcher:        SourceNamespaceMatcher,
		NamespaceLabelsSelector: SourceNamespaceLabelsSelector,
		PodLabelsSelector:       SourcePodLabelsSelector,
	}

	var sdaps []TrafficMatcher
	for _, ingress := range ingresses {
		sdaps = append(sdaps, BuildSourceDestAndPorts(family, policyNamespace, ingress.Ports, ingress.From))
	}
	return NewAny(sdaps...), DirectiveAllow
}

var EgressSelector = &SelectorFamily{}

func BuildTrafficPeersFromEgress(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) (TrafficMatcher, Directive) {
	if len(egresses) == 0 {
		// This is a hard-to-grok case:
		//   in network policy land, it should select NO Peers, with the effect of blocking all communication to target
		//   in matcher land, we translate that by matching ALL with a deny
		return EverythingMatcher, DirectiveDeny
	}

	// TODO this gets the job done but is ugly -- could it be better?
	family := &SelectorFamily{
		IPSelector:              DestinationIPSelector,
		IsExternalSelector:      DestinationIsExternalSelector,
		NamespaceMatcher:        DestinationNamespaceMatcher,
		NamespaceLabelsSelector: DestinationNamespaceLabelsSelector,
		PodLabelsSelector:       DestinationPodLabelsSelector,
	}

	var sdaps []TrafficMatcher
	for _, egress := range egresses {
		sdaps = append(sdaps, BuildSourceDestAndPorts(family, policyNamespace, egress.Ports, egress.To))
	}
	return NewAny(sdaps...), DirectiveAllow
}

type SelectorFamily struct {
	IPSelector              Selector
	IsExternalSelector      Selector
	NamespaceMatcher        func(string) *Equal
	NamespaceLabelsSelector Selector
	PodLabelsSelector       Selector
}

func BuildSourceDestAndPorts(family *SelectorFamily, policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) TrafficMatcher {
	sds := BuildSourceDestsFromSlice(family, policyNamespace, peers)
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

func BuildSourceDestsFromSlice(family *SelectorFamily, policyNamespace string, peers []networkingv1.NetworkPolicyPeer) TrafficMatcher {
	if len(peers) == 0 {
		return EverythingMatcher
	}
	var terms []TrafficMatcher
	for _, from := range peers {
		terms = append(terms, BuildSourceDest(family, policyNamespace, from))
	}
	return NewAny(terms...)
}

func isLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
}

func BuildSourceDest(family *SelectorFamily, policyNamespace string, peer networkingv1.NetworkPolicyPeer) TrafficMatcher {
	if peer.IPBlock != nil {
		return IPBlockMatcher(family.IPSelector, peer.IPBlock.CIDR, peer.IPBlock.Except)
	}
	podSel := peer.PodSelector
	nsSel := peer.NamespaceSelector
	isInternalMatcher := &Not{&Bool{family.IsExternalSelector}}
	if podSel == nil || isLabelSelectorEmpty(*podSel) {
		if nsSel == nil {
			return NewAll(
				isInternalMatcher,
				family.NamespaceMatcher(policyNamespace))
		} else if isLabelSelectorEmpty(*nsSel) {
			return isInternalMatcher
		} else {
			// nsSel has some stuff
			return NewAll(
				isInternalMatcher,
				KubeMatchLabelSelector(family.NamespaceLabelsSelector, *nsSel))
		}
	} else {
		podLabelsMatcher := KubeMatchLabelSelector(family.PodLabelsSelector, *podSel)
		// podSel has some stuff
		if nsSel == nil {
			return NewAll(
				isInternalMatcher,
				family.NamespaceMatcher(policyNamespace),
				podLabelsMatcher)
		} else if isLabelSelectorEmpty(*nsSel) {
			return NewAll(
				isInternalMatcher,
				KubeMatchLabelSelector(family.PodLabelsSelector, *podSel))
		} else {
			// nsSel has some stuff
			return NewAll(
				isInternalMatcher,
				KubeMatchLabelSelector(family.PodLabelsSelector, *podSel),
				KubeMatchLabelSelector(family.NamespaceLabelsSelector, *nsSel))
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
