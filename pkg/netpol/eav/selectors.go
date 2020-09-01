package eav

type Selector func(*Traffic) interface{}

func internalHelper(p *Peer, f func(ip *InternalPeer) interface{}) interface{} {
	if p.Internal == nil {
		return nil
	}
	return f(p.Internal)
}

func ConstantSelector(value interface{}) Selector {
	return func(traffic *Traffic) interface{} {
		return value
	}
}

var (
	ProtocolSelector Selector = func(traffic *Traffic) interface{} { return traffic.Protocol }
	PortSelector     Selector = func(traffic *Traffic) interface{} { return traffic.Port }

	SourceIPSelector         Selector = func(traffic *Traffic) interface{} { return traffic.Source.IP }
	SourceIsExternalSelector Selector = func(traffic *Traffic) interface{} { return traffic.Source.IsExternal() }
	SourceNamespaceSelector  Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Source, func(ip *InternalPeer) interface{} { return ip.Namespace })
	}
	SourcePodSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Source, func(ip *InternalPeer) interface{} { return ip.Pod })
	}
	SourceNodeSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Source, func(ip *InternalPeer) interface{} { return ip.Node })
	}
	SourceNamespaceLabelsSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Source, func(ip *InternalPeer) interface{} { return ip.NamespaceLabels })
	}
	SourcePodLabelsSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Source, func(ip *InternalPeer) interface{} { return ip.PodLabels })
	}
	SourceNodeLabelsSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Source, func(ip *InternalPeer) interface{} { return ip.NodeLabels })
	}

	DestinationIPSelector         Selector = func(traffic *Traffic) interface{} { return traffic.Destination.IP }
	DestinationIsExternalSelector Selector = func(traffic *Traffic) interface{} { return traffic.Destination.IsExternal() }
	DestinationNamespaceSelector  Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Destination, func(ip *InternalPeer) interface{} { return ip.Namespace })
	}
	DestinationPodSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Destination, func(ip *InternalPeer) interface{} { return ip.Pod })
	}
	DestinationNodeSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Destination, func(ip *InternalPeer) interface{} { return ip.Node })
	}
	DestinationNamespaceLabelsSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Destination, func(ip *InternalPeer) interface{} { return ip.NamespaceLabels })
	}
	DestinationPodLabelsSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Destination, func(ip *InternalPeer) interface{} { return ip.PodLabels })
	}
	DestinationNodeLabelsSelector Selector = func(traffic *Traffic) interface{} {
		return internalHelper(traffic.Destination, func(ip *InternalPeer) interface{} { return ip.NodeLabels })
	}
)
