package eav

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// All/All
var AllSourcesAllDests = &Policy{
	TrafficMatcher: EverythingMatcher,
	Directive:      DirectiveAllow,
}

// All/Internal
var AllSourcesInternalDests = &Policy{
	TrafficMatcher: &Not{&Bool{DestinationIsExternalSelector}},
	Directive:      DirectiveAllow,
}

// All/External
var AllSourcesExternalDests = &Policy{
	TrafficMatcher: &Bool{DestinationIsExternalSelector},
	Directive:      DirectiveAllow,
}

//// All/None -- this doesn't make any sense
//var AllSourcesNoDests = &Policy{
//	TrafficMatcher: NothingMatcher,
//	Directive:      DirectiveAllow,
//}

// Internal/All
var InternalSourcesAllDests = &Policy{
	TrafficMatcher: &Not{&Bool{SourceIsExternalSelector}},
	Directive:      DirectiveAllow,
}

// Internal/Internal
var InternalSourcesInternalDests = &Policy{
	TrafficMatcher: NewAll(
		&Not{&Bool{SourceIsExternalSelector}},
		&Not{&Bool{DestinationIsExternalSelector}}),
	Directive: DirectiveAllow,
}

// Internal/External
var InternalSourcesExternalDests = &Policy{
	TrafficMatcher: NewAll(
		&Not{&Bool{SourceIsExternalSelector}},
		&Bool{DestinationIsExternalSelector}),
	Directive: DirectiveAllow,
}

// Internal/None
//var InternalSourcesNoDests = &Policy{
//	TrafficMatcher: ???,
//	Directive:      DirectiveAllow,
//}

// TODO these are probably all useless -- since they don't match any targets:
// External/All
// External/Internal
// External/External
// External/None
// None/All
// None/Internal
// None/External
// None/None

var PodLabelSourceNamespaceLabelDest = &Policy{
	TrafficMatcher: NewAll(
		&LabelMatcher{
			Selector: SourcePodLabelsSelector,
			Key:      "app",
			Value:    "web",
		},
		&LabelMatcher{
			Selector: DestinationNamespaceLabelsSelector,
			Key:      "stage",
			Value:    "dev",
		}),
	Directive: DirectiveAllow,
}

var SameNamespaceSourceAndDest = &Policy{
	TrafficMatcher: SameNamespaceMatcher,
	Directive:      DirectiveAllow,
}

var (
	tcp     = v1.ProtocolTCP
	udp     = v1.ProtocolUDP
	port53  = intstr.FromInt(53)
	port80  = intstr.FromInt(80)
	port443 = intstr.FromInt(443)
	port988 = intstr.FromInt(988)
)

var AnthosAllowKubeDNSEgressNetworkPolicy = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-kube-dns-egress",
		Namespace: "kube-system",
		//Annotations:                nil,
		//#configmanagement.gke.io/cluster-selector: ${CLUSTER_SELECTOR}
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"k8s-app": "kube-dns",
			},
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &port53},
					{Protocol: &udp, Port: &port53},
				},
				To: []networkingv1.NetworkPolicyPeer{
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "169.254.169.254/32",
						},
					},
				},
			},
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &port443},
				},
				To: []networkingv1.NetworkPolicyPeer{
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "${APISERVER_IP}/32",
						},
					},
				},
			},
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &port443},
				},
				To: []networkingv1.NetworkPolicyPeer{
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "${GOOGLEAPIS_CIDR}/32",
						},
					},
				},
			},
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &port80},
				},
				To: []networkingv1.NetworkPolicyPeer{
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "169.254.169.254/32",
						},
					},
				},
			},
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &port988},
				},
				To: []networkingv1.NetworkPolicyPeer{
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "127.0.0.1/32",
						},
					},
				},
			},
		},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
	},
}

var AnthosAllowKubeDNSEgress = &Policy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-kube-dns-egress",
		Namespace: "kube-system",
		//Annotations:                nil,
		//#configmanagement.gke.io/cluster-selector: ${CLUSTER_SELECTOR}
	},
	TrafficMatcher: NewAll(
		SourceNamespaceMatcher("kube-system"),
		&LabelMatcher{
			Selector: SourceNamespaceLabelsSelector,
			Key:      "k8s-app",
			Value:    "kube-dns",
		},
		NewAny(
			NewAll(
				IPBlockMatcher(DestinationIPSelector, "169.254.169.254/32", []string{}),
				NewEqual(PortSelector, ConstantSelector(53)),
				NewAny(
					ProtocolMatcher(v1.ProtocolTCP),
					ProtocolMatcher(v1.ProtocolUDP),
				),
			),
			NewAll(
				IPBlockMatcher(DestinationIPSelector, "${APISERVER_IP}/32", []string{}),
				NewEqual(PortSelector, ConstantSelector(443)),
				ProtocolMatcher(v1.ProtocolTCP),
			),
			NewAll(
				IPBlockMatcher(DestinationIPSelector, "${GOOGLEAPIS_CIDR}", []string{}),
				NewEqual(PortSelector, ConstantSelector(443)),
				ProtocolMatcher(v1.ProtocolTCP),
			),
			NewAll(
				IPBlockMatcher(DestinationIPSelector, "169.254.169.254/32", []string{}),
				NewEqual(PortSelector, ConstantSelector(80)),
				ProtocolMatcher(v1.ProtocolTCP),
			),
			NewAll(
				IPBlockMatcher(DestinationIPSelector, "127.0.0.1/32", []string{}),
				NewEqual(PortSelector, ConstantSelector(988)),
				ProtocolMatcher(v1.ProtocolTCP),
			),
		),
	),
	Directive: DirectiveAllow,
}
