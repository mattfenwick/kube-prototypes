package matcher

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PortMatcher interface {
	Allows(port *PortProtocol) bool
}

// AllPortsAllProtocols models the case where no ports/protocols are
// specified, which is treated as "allow any" by NetworkPolicy
type AllPortsAllProtocols struct{}

func (ap *AllPortsAllProtocols) Allows(pp *PortProtocol) bool {
	return true
}

func (ap *AllPortsAllProtocols) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]string{
		"Type": "all ports all protocols",
	})
}

// AllPortsOnProtocol models the case where a protocol is specified but
// a port number/name is not, which is treated as "allow any number/named
// port on the matching protocol"
type AllPortsOnProtocol struct {
	Protocol v1.Protocol
}

func (apop *AllPortsOnProtocol) Allows(pp *PortProtocol) bool {
	return apop.Protocol == pp.Protocol
}

func (apop *AllPortsOnProtocol) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "all ports on protocol",
		"Protocol": apop.Protocol,
	})
}

// ExactPortProtocol models the case where traffic must match a protocol and
// a number/named port
type ExactPortProtocol struct {
	Protocol v1.Protocol
	Port     intstr.IntOrString
}

func (epp *ExactPortProtocol) Allows(other *PortProtocol) bool {
	return other.Protocol == epp.Protocol && isPortMatch(other.Port, epp.Port)
}

func (epp *ExactPortProtocol) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "port on protocol",
		"Protocol": epp.Protocol,
		"Port":     epp.Port,
	})
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
