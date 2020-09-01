package eav

import v1 "k8s.io/api/core/v1"

type Blackduck struct {
	Namespace string
	KBAddress string
}

func (bd *Blackduck) DenyAll() *Policy {
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
	return &Policy{
		TrafficMatcher: NewAny(
			NewEqual(
				SourceNamespaceSelector,
				ConstantSelector(bd.Namespace)),
			NumberedPortMatcher(53),
			ProtocolMatcher(v1.ProtocolTCP)),
		Directive: DirectiveAllow,
	}
}

func (bd *Blackduck) AllowEgressToKB() *Policy {
	return &Policy{
		TrafficMatcher: NewAny(
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
	return &Policy{
		TrafficMatcher: NewAny(
			NewEqual(
				SourceNamespaceSelector,
				ConstantSelector(bd.Namespace)),
			NewEqual(
				DestinationNamespaceSelector,
				ConstantSelector(bd.Namespace))),
		Directive: DirectiveAllow,
	}
}
