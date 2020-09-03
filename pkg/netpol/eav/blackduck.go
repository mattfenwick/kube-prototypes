package eav

import v1 "k8s.io/api/core/v1"

type Blackduck struct {
	Namespace string
	KBAddress string
}

func (bd *Blackduck) DenyAll() *Policy {
	// Explanation: blanket (TODO: low priority) deny of *everything* to and from Blackduck
	return &Policy{
		TrafficMatcher: NewAny(
			NewEqual(
				SourceNamespaceSelector,
				ConstantSelector(bd.Namespace)),
			NewEqual(
				DestinationNamespaceSelector,
				ConstantSelector(bd.Namespace))),
		Directive: DirectiveDeny,
	}
}

func (bd *Blackduck) AllowDNSOnTCP() *Policy {
	// Explanation: allow DNS from the Blackduck namespace
	return &Policy{
		TrafficMatcher: NewAll(
			NewEqual(
				SourceNamespaceSelector,
				ConstantSelector(bd.Namespace)),
			NumberedPortMatcher(53),
			ProtocolMatcher(v1.ProtocolTCP)),
		Directive: DirectiveAllow,
	}
}

func (bd *Blackduck) AllowEgressToKB() *Policy {
	// Explanation: if source is the Blackduck namespace, and the destination
	// is the KB API, allow the traffic
	return &Policy{
		TrafficMatcher: NewAll(
			NewEqual(
				SourceNamespaceSelector,
				ConstantSelector(bd.Namespace)),
			NewEqual(
				DestinationIPSelector,
				ConstantSelector(bd.KBAddress))),
		Directive: DirectiveAllow,
	}
}

func (bd *Blackduck) AllowBDNamespaceCommunication() *Policy {
	// Explanation: if both source and destination are the Blackduck namespace,
	// allow the traffic
	return &Policy{
		TrafficMatcher: NewAll(
			NewEqual(
				SourceNamespaceSelector,
				ConstantSelector(bd.Namespace)),
			NewEqual(
				DestinationNamespaceSelector,
				ConstantSelector(bd.Namespace))),
		Directive: DirectiveAllow,
	}
}
