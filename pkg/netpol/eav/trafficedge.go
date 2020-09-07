package eav

import (
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/kube"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type TrafficMatchType string

const (
	TrafficMatchTypeAll TrafficMatchType = "all"
	TrafficMatchTypeAny TrafficMatchType = "any"
)

type TrafficEdge struct {
	Type     TrafficMatchType
	Source   *PeerMatcher
	Dest     *PeerMatcher
	Port     *PortMatcher
	Protocol *ProtocolMatcher
}

func (m *TrafficEdge) Matches(t *Traffic) bool {
	switch m.Type {
	case TrafficMatchTypeAll:
		if m.Source != nil && !m.Source.Matches(t.Source) {
			return false
		}
		if m.Dest != nil && !m.Dest.Matches(t.Destination) {
			return false
		}
		if m.Port != nil && !m.Port.Matches(t.Port) {
			return false
		}
		if m.Protocol != nil && !m.Protocol.Matches(t.Protocol) {
			return false
		}
		return true
	case TrafficMatchTypeAny:
		if m.Source != nil && m.Source.Matches(t.Source) {
			return true
		}
		if m.Dest != nil && m.Dest.Matches(t.Destination) {
			return true
		}
		if m.Port != nil && m.Port.Matches(t.Port) {
			return true
		}
		if m.Protocol != nil && m.Protocol.Matches(t.Protocol) {
			return true
		}
		return false
	default:
		panic(errors.Errorf("invalid match type %s", m.Type))
	}
}

type PeerLocation string

var (
	PeerLocationInternal PeerLocation = "internal"
	PeerLocationExternal PeerLocation = "external"
)

type PeerMatcher struct {
	IP               *IPMatcher
	RelativeLocation *PeerLocation
	Internal         *InternalPeerMatcher
}

func (pm *PeerMatcher) Matches(p *Peer) bool {
	if pm.IP != nil && !pm.IP.Matches(p.IP) {
		return false
	}
	if pm.RelativeLocation != nil {
		if *pm.RelativeLocation == PeerLocationInternal && p.IsExternal() {
			return false
		}
		if *pm.RelativeLocation == PeerLocationExternal && !p.IsExternal() {
			return false
		}
	}
	if pm.Internal != nil && !pm.Internal.Matches(p.Internal) {
		return false
	}
	return true
}

type InternalPeerMatcher struct {
	Namespace       *StringMatcher
	NamespaceLabels *metav1.LabelSelector
	Node            *StringMatcher
	NodeLabels      *metav1.LabelSelector
	Pod             *StringMatcher
	PodLabels       *metav1.LabelSelector
	// TODO services, deployments?
	//Service *StringMatcher
}

func (ipm *InternalPeerMatcher) Matches(i *InternalPeer) bool {
	if ipm.Namespace != nil && !ipm.Namespace.Matches(i.Namespace) {
		return false
	}
	if ipm.NamespaceLabels != nil && !kube.IsLabelsMatchLabelSelector(i.NamespaceLabels, *ipm.NamespaceLabels) {
		return false
	}
	if ipm.Node != nil && !ipm.Node.Matches(i.Node) {
		return false
	}
	if ipm.NodeLabels != nil && !kube.IsLabelsMatchLabelSelector(i.NodeLabels, *ipm.NodeLabels) {
		return false
	}
	if ipm.Pod != nil && !ipm.Pod.Matches(i.Pod) {
		return false
	}
	if ipm.PodLabels != nil && !kube.IsLabelsMatchLabelSelector(i.PodLabels, *ipm.PodLabels) {
		return false
	}
	return true
}

type StringMatcher struct {
	Value string
}

func (sm *StringMatcher) Matches(v string) bool {
	return sm.Value == v
}

type PortMatcher struct {
	Range *struct {
		Low  int
		High int
	}
	Value *intstr.IntOrString
}

func (pm *PortMatcher) Matches(port intstr.IntOrString) bool {
	if (pm.Range == nil && pm.Value == nil) || (pm.Range != nil && pm.Value != nil) {
		panic("either Range or Value must be specified")
	}
	if pm.Range != nil && port.Type == intstr.Int {
		portNumber := int(port.IntVal)
		return portNumber >= pm.Range.Low && portNumber < pm.Range.High
	}
	return isPortMatch(port, *pm.Value)
}

func isPortMatch(a intstr.IntOrString, b intstr.IntOrString) bool {
	switch a.Type {
	case intstr.Int:
		switch b.Type {
		case intstr.Int:
			return a.IntVal == b.IntVal
		case intstr.String:
			// TODO what if this named port resolves to same int?
			return false
		default:
			panic("invalid type")
		}
	case intstr.String:
		switch b.Type {
		case intstr.Int:
			// TODO what if this named port resolves to same int?
			return false
		case intstr.String:
			return a.StrVal == b.StrVal
		default:
			panic("invalid type")
		}
	default:
		panic("invalid type")
	}
}

type ProtocolMatcher struct {
	Values []v1.Protocol
}

func (pm *ProtocolMatcher) Matches(protocol v1.Protocol) bool {
	for _, prot := range pm.Values {
		if prot == protocol {
			return true
		}
	}
	return false
}

// NothingMatcher matches nothing
var NothingMatcher = &TrafficEdge{Type: TrafficMatchTypeAny}

// EverythingMatcher matches everything
var EverythingMatcher = &TrafficEdge{Type: TrafficMatchTypeAll}

func NamedPortMatcher(port string) *PortMatcher {
	portRef := intstr.FromString(port)
	return &PortMatcher{
		Value: &portRef,
	}
}

func NumberedPortMatcher(port int) *PortMatcher {
	portRef := intstr.FromInt(port)
	return &PortMatcher{Value: &portRef}
}

// IPMatcher matches an IP address using a cidr
type IPMatcher struct {
	Value *string
	Block *networkingv1.IPBlock
}

func (ipm *IPMatcher) Matches(ip string) bool {
	if ipm.Value != nil {
		return *ipm.Value == ip
	}
	return kube.IsIPBlockMatchForIP(ip, ipm.Block)
}
