package obsolete

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var AnyProtocolPortMatcher = &ProtocolPortMatcher{
	PortMatcher:     &AnyPortMatcher{},
	ProtocolMatcher: &AnyProtocolMatcher{},
}

type ProtocolPortMatcher struct {
	PortMatcher     PortMatcher
	ProtocolMatcher ProtocolMatcher
}

func (pm *ProtocolPortMatcher) IsProtocolPortMatch(protocol v1.Protocol, port intstr.IntOrString) bool {
	return pm.ProtocolMatcher.IsProtocolMatch(protocol) && pm.PortMatcher.IsPortMatch(port)
}

type ProtocolMatcher interface {
	IsProtocolMatch(v1.Protocol) bool
}

type AnyProtocolMatcher struct{}

func (apm *AnyProtocolMatcher) IsProtocolMatch(protocol v1.Protocol) bool {
	return true
}

type SpecifiedProtocolMatcher struct {
	Protocols []v1.Protocol
}

func (spm *SpecifiedProtocolMatcher) IsProtocolMatch(protocol v1.Protocol) bool {
	if len(spm.Protocols) == 0 {
		panic("SpecifiedProtocolMatcher must have >= 1 protocol")
	}
	for _, prot := range spm.Protocols {
		if prot == protocol {
			return true
		}
	}
	return false
}

type PortMatcher interface {
	IsPortMatch(intstr.IntOrString) bool
}

// AnyPortMatcher matches any port
type AnyPortMatcher struct{}

func (apm *AnyPortMatcher) IsPortMatch(port intstr.IntOrString) bool {
	return true
}

// RangePortMatcher implements a port range of [Low, High)
// Thus, the lower bound IS included in the range, and the upper
// bound is NOT included in the range.
type RangePortMatcher struct {
	Low  int
	High int
}

func (rpm *RangePortMatcher) IsPortMatch(port intstr.IntOrString) bool {
	if port.Type != intstr.Int {
		return false
	}
	portNumber := int(port.IntVal)
	return portNumber >= rpm.Low && portNumber < rpm.High
}

type SpecifiedPortsPortMatcher struct {
	Named    []string
	Numbered []int
}

func (sppm *SpecifiedPortsPortMatcher) IsPortMatch(port intstr.IntOrString) bool {
	if len(sppm.Named) == 0 && len(sppm.Numbered) == 0 {
		panic("SpecifiedPortsPortMatcher must have >= 1 named or numbered ports")
	}
	switch port.Type {
	case intstr.Int:
		numberedPort := int(port.IntVal)
		for _, p := range sppm.Numbered {
			if p == numberedPort {
				return true
			}
		}
		return false
	case intstr.String:
		namedPort := port.StrVal
		for _, p := range sppm.Named {
			if p == namedPort {
				return true
			}
		}
		return false
	default:
		panic("invalid intstr Type")
	}
}
