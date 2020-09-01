package eav

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type TrafficMatcher interface {
	Matches(t *Traffic) bool
}

// NothingMatcher matches nothing
var NothingMatcher = NewAny()

// EverythingMatcher matches everything
var EverythingMatcher = NewAll()

// AnyInternalMatcher matches any pod in the same kube clusters
var AnyInternalMatcher = NewEqual(SourceIsExternalSelector, DestinationIsExternalSelector, ConstantSelector(false))

// AnyExternalMatcher matches anything NOT in the same kube cluster
var AnyExternalMatcher = NewEqual(SourceIsExternalSelector, DestinationIsExternalSelector, ConstantSelector(true))

// SourceNamespaceMatcher matches Traffic whose source is internal and has a namespace of ns
func SourceNamespaceMatcher(ns string) *Equal {
	return NewEqual(SourceNamespaceSelector, ConstantSelector(ns))
}

// DestinationNamespaceMatcher matches Traffic whose destination is internal and has a namespace of ns
func DestinationNamespaceMatcher(ns string) *Equal {
	return NewEqual(DestinationNamespaceSelector, ConstantSelector(ns))
}

// SourceNodeMatcher ...
func SourceNodeMatcher(node string) *Equal {
	return NewEqual(SourceNodeSelector, ConstantSelector(node))
}

// DestinationNodeMatcher ...
func DestinationNodeMatcher(node string) *Equal {
	return NewEqual(DestinationNodeSelector, ConstantSelector(node))
}

// SourcePodMatcher ...
func SourcePodMatcher(pod string) *Equal {
	return NewEqual(SourcePodSelector, ConstantSelector(pod))
}

// DestinationPodMatcher ...
func DestinationPodMatcher(pod string) *Equal {
	return NewEqual(DestinationPodSelector, ConstantSelector(pod))
}

var SameNamespaceMatcher = NewEqual(SourceNamespaceSelector, DestinationNamespaceSelector)
var SameNodeMatcher = NewEqual(SourceNodeSelector, DestinationNodeSelector)

func NamedPortMatcher(port string) TrafficMatcher {
	return NewEqual(PortSelector, ConstantSelector(port))
}

func NumberedPortMatcher(port int) TrafficMatcher {
	return NewEqual(PortSelector, ConstantSelector(port))
}

func ProtocolMatcher(protocol v1.Protocol) TrafficMatcher {
	return NewEqual(ProtocolSelector, ConstantSelector(protocol))
}

// RangePortMatcher implements a port range of [Low, High)
// Thus, the lower bound IS included in the range, and the upper
// bound is NOT included in the range.
type RangePortMatcher struct {
	Low  int
	High int
}

func (rpm *RangePortMatcher) Matches(t *Traffic) bool {
	port := t.Port
	if port.Type != intstr.Int {
		return false
	}
	portNumber := int(port.IntVal)
	return portNumber >= rpm.Low && portNumber < rpm.High
}
