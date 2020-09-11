package crd

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO wow, this is really hard due to the target-biased nature of network
//   policies

func AllowAllIngressNetworkingPolicy(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-all",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

func AllowAllEgressNetworkingPolicy(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-all",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

func ReduceAll(np []*Policy) []*networkingv1.NetworkPolicy {
	var netpols []*networkingv1.NetworkPolicy
	for _, n := range np {
		netpols = append(netpols, Reduce(n)...)
	}
	return append(netpols)
}

func Reduce(np *Policy) []*networkingv1.NetworkPolicy {
	var pols []*networkingv1.NetworkPolicy
	for i, policyType := range np.Spec.Compatibility {
		switch policyType {
		case networkingv1.PolicyTypeIngress:
			ingress, _ := ReduceRules(true, np.Spec.TrafficMatcher)
			namespace := "default"
			edge := np.Spec.TrafficMatcher
			podSelector := metav1.LabelSelector{}
			if edge.Dest != nil {
				namespace = edge.Dest.Internal.Namespace.Value
				podSelector = *edge.Dest.Internal.PodLabels
			}
			pols = append(pols, &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-ingress-%d", np.Name, i),
					Namespace: namespace,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: podSelector,
					Ingress:     ingress,
					PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
				},
			})
		case networkingv1.PolicyTypeEgress:
			_, egress := ReduceRules(false, np.Spec.TrafficMatcher)
			namespace := "default"
			edge := np.Spec.TrafficMatcher
			podSelector := metav1.LabelSelector{}
			if edge.Source != nil {
				if edge.Source.Internal.Namespace != nil {
					namespace = edge.Source.Internal.Namespace.Value
				}
				if edge.Source.Internal.PodLabels != nil {
					podSelector = *edge.Source.Internal.PodLabels
				}
			}
			pols = append(pols, &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-egress-%d", np.Name, i),
					Namespace: namespace,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: podSelector,
					Egress:      egress,
					PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
				},
			})
		}
	}
	return pols
}

func ReduceRules(isIngress bool, edge *TrafficEdge) ([]networkingv1.NetworkPolicyIngressRule, []networkingv1.NetworkPolicyEgressRule) {
	var peers []networkingv1.NetworkPolicyPeer
	if isIngress {
		peers = ReducePeerMatcher(edge.Source)
	} else {
		peers = ReducePeerMatcher(edge.Dest)
	}
	ports := ReducePortProtocol(edge.Port, edge.Protocol)
	var ingress []networkingv1.NetworkPolicyIngressRule
	var egress []networkingv1.NetworkPolicyEgressRule
	if isIngress {
		ingress = append(ingress, networkingv1.NetworkPolicyIngressRule{
			Ports: ports,
			From:  peers,
		})
	} else {
		egress = append(egress, networkingv1.NetworkPolicyEgressRule{
			Ports: ports,
			To:    peers,
		})
	}
	return ingress, egress
}

func ReducePortProtocol(portMatcher *PortMatcher, protocolMatcher *ProtocolMatcher) []networkingv1.NetworkPolicyPort {
	var protocols []v1.Protocol
	if protocolMatcher != nil {
		for _, p := range protocolMatcher.Values {
			protocols = append(protocols, p)
		}
	}
	if len(protocols) == 0 {
		protocols = []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP} // TODO I guess SCTP can't be used?, v1.ProtocolSCTP}
	}
	var npPorts []networkingv1.NetworkPolicyPort
	if portMatcher == nil {
		for _, protocol := range protocols {
			protocolRef := protocol
			npPorts = append(npPorts, networkingv1.NetworkPolicyPort{
				Protocol: &protocolRef,
			})
		}
	} else if portMatcher.Value != nil {
		for _, protocol := range protocols {
			// so that we don't get a ref to the wrong variable
			protocolRef := protocol
			npPorts = append(npPorts, networkingv1.NetworkPolicyPort{
				Protocol: &protocolRef,
				Port:     portMatcher.Value,
			})
		}
	} else if portMatcher.Range != nil {
		for _, protocol := range protocols {
			// so that we don't get a ref to the wrong variable
			protocolRef := protocol
			for port := portMatcher.Range.Low; port < portMatcher.Range.High; port++ {
				portRef := intstr.FromInt(port)
				npPorts = append(npPorts, networkingv1.NetworkPolicyPort{
					Protocol: &protocolRef,
					Port:     &portRef,
				})
			}
		}
	}
	return npPorts
}

func ReducePeerMatcher(peer *PeerMatcher) []networkingv1.NetworkPolicyPeer {
	// TODO check any/all of policy spec?
	if peer == nil {
		// TODO what should this return?  for now we'll just match everything
		return []networkingv1.NetworkPolicyPeer{}
	}
	if peer.IP != nil {
		if peer.IP.Value != nil {
			panic("unable to translate IP value")
		}
		return []networkingv1.NetworkPolicyPeer{{IPBlock: peer.IP.Block}}
	} else if peer.Internal != nil {
		i := peer.Internal
		if i.Pod != nil || i.Namespace != nil || i.NodeLabels != nil || i.Node != nil {
			panic("unimplemented feature: internal peer matcher (pod/namespace/node/nodelabels")
		}
		peer := networkingv1.NetworkPolicyPeer{}
		if i.PodLabels != nil {
			peer.PodSelector = i.PodLabels
		}
		if i.NamespaceLabels != nil {
			peer.NamespaceSelector = i.NamespaceLabels
		}
		return []networkingv1.NetworkPolicyPeer{peer}
	}
	return []networkingv1.NetworkPolicyPeer{}
}
