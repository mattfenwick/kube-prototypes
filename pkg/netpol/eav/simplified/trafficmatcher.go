package eav

import (
	"github.com/mattfenwick/kube-prototypes/pkg/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TrafficMatcher interface {
	Matches(tm *TrafficMap) bool
}

// NothingMatcher matches nothing
var NothingMatcher = NewAny()

// EverythingMatcher matches everything
var EverythingMatcher = NewAll()

func isExternalMatcher(rootPath string) TrafficMatcher {
	return NewEqual(NewKeyPathSelector(rootPath, InternalSelector), &ConstantSelector{nil})
}

var (
	SourceIsExternalMatcher = isExternalMatcher(SourceSelector)
	SourceIsInternalMatcher = &Not{SourceIsExternalMatcher}

	DestIsExternalMatcher = isExternalMatcher(DestSelector)
	DestIsInternalMatcher = &Not{DestIsExternalMatcher}
)

// AnyInternalMatcher matches communication between two things in the same kube cluster
var AnyInternalMatcher = NewAll(SourceIsInternalMatcher, DestIsInternalMatcher)

// AnyExternalMatcher matches anything NOT in the same kube cluster
var AnyExternalMatcher = NewAny(SourceIsExternalMatcher, DestIsExternalMatcher)

// SourceNamespaceMatcher matches Traffic whose source is internal and has a namespace of ns
func SourceNamespaceMatcher(ns string) *Equal {
	return NewEqual(NewKeyPathSelector(SourceSelector, InternalSelector, NamespaceSelector), &ConstantSelector{ns})
}

// DestNamespaceMatcher matches Traffic whose Dest is internal and has a namespace of ns
func DestNamespaceMatcher(ns string) *Equal {
	return NewEqual(NewKeyPathSelector(DestSelector, InternalSelector, NamespaceSelector), &ConstantSelector{ns})
}

// SourceNodeMatcher ...
func SourceNodeMatcher(node string) *Equal {
	return NewEqual(NewKeyPathSelector(SourceSelector, InternalSelector, NodeSelector), &ConstantSelector{node})
}

// DestNodeMatcher ...
func DestNodeMatcher(node string) *Equal {
	return NewEqual(NewKeyPathSelector(DestSelector, InternalSelector, NodeSelector), &ConstantSelector{node})
}

// SourcePodMatcher ...
func SourcePodMatcher(pod string) *Equal {
	return NewEqual(NewKeyPathSelector(SourceSelector, InternalSelector, PodSelector), &ConstantSelector{pod})
}

// DestPodMatcher ...
func DestPodMatcher(pod string) *Equal {
	return NewEqual(NewKeyPathSelector(DestSelector, InternalSelector, PodSelector), &ConstantSelector{pod})
}

var SameNamespaceMatcher = NewEqual(
	NewKeyPathSelector(SourceSelector, InternalSelector, NamespaceSelector),
	NewKeyPathSelector(DestSelector, InternalSelector, NamespaceSelector))
var SameNodeMatcher = NewEqual(
	NewKeyPathSelector(SourceSelector, InternalSelector, NodeSelector),
	NewKeyPathSelector(DestSelector, InternalSelector, NodeSelector))

func NamedPortMatcher(port string) TrafficMatcher {
	return NewEqual(NewKeyPathSelector(PortSelector), &ConstantSelector{port})
}

func NumberedPortMatcher(port int) TrafficMatcher {
	return NewEqual(NewKeyPathSelector(PortSelector), &ConstantSelector{port})
}

func ProtocolMatcher(protocol v1.Protocol) TrafficMatcher {
	return NewEqual(NewKeyPathSelector(ProtocolSelector), &ConstantSelector{protocol})
}

// RangePortMatcher implements a port range of [Low, High)
// Thus, the lower bound IS included in the range, and the upper
// bound is NOT included in the range.
type RangePortMatcher struct {
	Low  int
	High int
}

func (rpm *RangePortMatcher) Matches(tm *TrafficMap) bool {
	portNumber, ok := NewKeyPathSelector(PortSelector).Select(tm).(int)
	if !ok {
		return false
	}
	return portNumber >= rpm.Low && portNumber < rpm.High
}

func KubeMatchLabels(selector Selector, labels map[string]string) TrafficMatcher {
	var terms []TrafficMatcher
	for key, val := range labels {
		terms = append(terms, &LabelMatcher{
			Selector: selector,
			Key:      key,
			Value:    val,
		})
	}
	return NewAll(terms...)
}

type KubeMatchExpressionMatcher struct {
	Selector   Selector
	Expression metav1.LabelSelectorRequirement
}

func (kmem *KubeMatchExpressionMatcher) Matches(tm *TrafficMap) bool {
	labels := kmem.Selector.Select(tm).(map[string]string)
	return kube.IsMatchExpressionMatchForLabels(labels, kmem.Expression)
}

func KubeMatchExpressions(selector Selector, mes []metav1.LabelSelectorRequirement) TrafficMatcher {
	var terms []TrafficMatcher
	for _, exp := range mes {
		terms = append(terms, &KubeMatchExpressionMatcher{
			Selector:   selector,
			Expression: exp,
		})
	}
	return NewAll(terms...)
}

func KubeMatchLabelSelector(selector Selector, ls metav1.LabelSelector) TrafficMatcher {
	return NewAll(
		KubeMatchLabels(selector, ls.MatchLabels),
		KubeMatchExpressions(selector, ls.MatchExpressions))
}

type LabelMatcher struct {
	Selector Selector
	Key      string
	Value    string
}

func (lm *LabelMatcher) Matches(tm *TrafficMap) bool {
	labels := lm.Selector.Select(tm).(map[string]string)
	value, ok := labels[lm.Key]
	return ok && value == lm.Value
}

// IPMatcher matches an IP address using a cidr
type IPMatcher struct {
	Selector Selector
	CIDR     string
}

func (ipm *IPMatcher) Matches(tm *TrafficMap) bool {
	ip := ipm.Selector.Select(tm).(string)
	return kube.IsIPInCIDR(ip, ipm.CIDR)
}

func IPBlockMatcher(selector Selector, cidr string, except []string) TrafficMatcher {
	// TODO wow this is unpleasant, is there a way to not have to do this?
	var values []interface{}
	for _, e := range except {
		values = append(values, e)
	}
	return NewAll(
		&IPMatcher{Selector: selector, CIDR: cidr},
		&Not{Term: &InArray{
			Selector: selector,
			Values:   values}})
}
