package garbage

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Garbage_Traffic represents a request from or to a target's source/dest counterpart
type Garbage_Traffic struct {
	Counterpart Garbage_TrafficCounterpart

	// ResolvedTarget is the object of a network policy -- the ??pod?? that
	//   is potentially issuing an egress or receiving an ingress
	// It sounds like it's not possible to make network policies targeting services
	//   unless you think of resolving services down to pods and adding those into
	//   iptables -- which CNIs may do
	ResolvedTarget Garbage_Target
}

type Garbage_TrafficCounterpart struct {
	// InternalSourceDest is the counterpart that's communicating with Peer.
	//   If this is a pod in the same cluster, gather up information about that
	//   pod -- labels, namespace, etc.  Otherwise, use nil for this field which
	//   will be interpreted as 'External'.
	InternalSourceDest *struct {
		PodLabels       map[string]string
		NamespaceLabels map[string]string
		Namespace       string
	}

	IsIngress bool
	Protocol  v1.Protocol
	Port      intstr.IntOrString
	IP        string
}

type Garbage_Target struct {
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
}

func (tc *Garbage_TrafficCounterpart) IsExternal() bool {
	return tc.InternalSourceDest == nil
}

// NetworkPolicyRule models a rule for matching a Peer and/or Counterpart and/or Traffic
type NetworkPolicyRule struct {
	// TODO could combine into one single `func(Traffic) bool` matcher
	//   could also split into Peer/Counterpart/Traffic matcher
	// TODO can't serialize arbitrary functions -- need to model matchers as data
	//
	TargetMatcher      func(Garbage_Target) bool
	CounterpartMatcher func(Garbage_TrafficCounterpart) bool
}
