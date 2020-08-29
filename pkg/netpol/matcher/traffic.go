package matcher

// Traffic represents a request from or to a target's source/dest counterpart
type Traffic struct {
	IsExternal      bool
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
	IsIngress       bool
	Port            *PortProtocol
	IP              string
}

func NewPodTraffic(podLabels map[string]string, nsLabels map[string]string, ns string, isIngress bool, port *PortProtocol, ip string) *Traffic {
	return &Traffic{
		IsExternal:      false,
		PodLabels:       podLabels,
		NamespaceLabels: nsLabels,
		Namespace:       ns,
		IsIngress:       isIngress,
		Port:            port,
		IP:              ip,
	}
}

func NewExternalTraffic(isIngress bool, port *PortProtocol, ip string) *Traffic {
	return &Traffic{
		IsExternal:      true,
		PodLabels:       nil,
		NamespaceLabels: nil,
		Namespace:       "",
		IsIngress:       isIngress,
		Port:            port,
		IP:              ip,
	}
}
