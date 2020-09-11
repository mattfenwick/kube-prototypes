package crd

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Denies

var DenyAll = &Policy{
	ObjectMeta: metav1.ObjectMeta{
		Name: "deny-all",
	},
	Spec: PolicySpec{
		Compatibility:  []networkingv1.PolicyType{networkingv1.PolicyTypeEgress, networkingv1.PolicyTypeIngress},
		TrafficMatcher: EverythingMatcher,
		Directive:      DirectiveDeny,
	},
}

// All/All
var AllSourcesAllDests = &Policy{
	ObjectMeta: metav1.ObjectMeta{
		Name: "allow-all-sources-all-dests",
	},
	Spec: PolicySpec{
		Compatibility:  []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
		TrafficMatcher: EverythingMatcher,
		Directive:      DirectiveAllow,
	},
}

// All/Internal
var AllSourcesInternalDests = &Policy{
	ObjectMeta: metav1.ObjectMeta{
		Name: "all-sources-internal-dests",
	},
	Spec: PolicySpec{
		Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		TrafficMatcher: &TrafficEdge{
			Type: TrafficMatchTypeAll,
			Dest: &PeerMatcher{
				RelativeLocation: &PeerLocationInternal,
			},
		},
		Directive: DirectiveAllow,
	},
}

// All/External
var AllSourcesExternalDests = &Policy{
	ObjectMeta: metav1.ObjectMeta{
		Name: "all-sources-external-dests",
	},
	Spec: PolicySpec{
		Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		TrafficMatcher: &TrafficEdge{
			Type: TrafficMatchTypeAll,
			Dest: &PeerMatcher{
				RelativeLocation: &PeerLocationExternal,
			},
		},
		Directive: DirectiveAllow,
	},
}

//// All/None -- this doesn't make any sense
//var AllSourcesNoDests = &Policy{
//	TrafficEdge: NothingMatcher,
//	Directive:      DirectiveAllow,
//}

//// Internal/All
//var InternalSourcesAllDests = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name: "internal-sources-all-dests",
//	},
//	Spec: PolicySpec{
//		TrafficMatcher: SourceIsInternalMatcher,
//		Directive:      DirectiveAllow,
//	},
//}

//// Internal/Internal
//var InternalSourcesInternalDests = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name: "internal-sources-internal-dests",
//	},
//	Spec: PolicySpec{
//		TrafficMatcher: NewAll(
//			SourceIsInternalMatcher,
//			DestIsInternalMatcher),
//		Directive: DirectiveAllow,
//	},
//}
//
//// Internal/External
//var InternalSourcesExternalDests = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name: "internal-sources-external-dests",
//	},
//	Spec: PolicySpec{
//		TrafficMatcher: NewAll(
//			SourceIsInternalMatcher,
//			DestIsExternalMatcher),
//		Directive: DirectiveAllow,
//	},
//}

// Internal/None
//var InternalSourcesNoDests = &Policy{
//	TrafficEdge: ???,
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

//var PodLabelSourceNamespaceLabelDest = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name: "pod-label-source-namespace-label-dest",
//	},
//	Spec: PolicySpec{
//		TrafficMatcher: NewAll(
//			&LabelMatcher{
//				Selector: NewKeyPathSelector(SourceSelector, InternalSelector, PodLabelsSelector),
//				Key:      "app",
//				Value:    "web",
//			},
//			&LabelMatcher{
//				Selector: NewKeyPathSelector(DestSelector, InternalSelector, NamespaceLabelsSelector),
//				Key:      "stage",
//				Value:    "dev",
//			}),
//		Directive: DirectiveAllow,
//	},
//}
//
//var SameNamespaceSourceAndDest = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name: "same-namespace-source-and-dest",
//	},
//	Spec: PolicySpec{
//		TrafficMatcher: SameNamespaceMatcher,
//		Directive:      DirectiveAllow,
//	},
//}

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

//var AnthosAllowKubeDNSEgress = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name:      "allow-kube-dns-egress",
//		Namespace: "kube-system",
//		//Annotations:                nil,
//		//#configmanagement.gke.io/cluster-selector: ${CLUSTER_SELECTOR}
//	},
//	Spec: PolicySpec{
//		Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
//		TrafficMatcher: NewAll(
//			SourceNamespaceMatcher("kube-system"),
//			&LabelMatcher{
//				Selector: NewKeyPathSelector(SourceSelector, InternalSelector, NamespaceLabelsSelector),
//				Key:      "k8s-app",
//				Value:    "kube-dns",
//			},
//			NewAny(
//				NewAll(
//					IPBlockMatcher(NewKeyPathSelector(DestSelector, IPSelector), "169.254.169.254/32", []string{}),
//					NumberedPortMatcher(53),
//					NewAny(
//						ProtocolMatcher(v1.ProtocolTCP),
//						ProtocolMatcher(v1.ProtocolUDP),
//					),
//				),
//				NewAll(
//					IPBlockMatcher(NewKeyPathSelector(DestSelector, IPSelector), "${APISERVER_IP}/32", []string{}),
//					NumberedPortMatcher(443),
//					ProtocolMatcher(v1.ProtocolTCP),
//				),
//				NewAll(
//					IPBlockMatcher(NewKeyPathSelector(DestSelector, IPSelector), "${GOOGLEAPIS_CIDR}", []string{}),
//					NumberedPortMatcher(443),
//					ProtocolMatcher(v1.ProtocolTCP),
//				),
//				NewAll(
//					IPBlockMatcher(NewKeyPathSelector(DestSelector, IPSelector), "169.254.169.254/32", []string{}),
//					NumberedPortMatcher(80),
//					ProtocolMatcher(v1.ProtocolTCP),
//				),
//				NewAll(
//					IPBlockMatcher(NewKeyPathSelector(DestSelector, IPSelector), "127.0.0.1/32", []string{}),
//					NumberedPortMatcher(988),
//					ProtocolMatcher(v1.ProtocolTCP),
//				),
//			),
//		),
//		Directive: DirectiveAllow,
//	},
//}

// See: https://github.com/GoogleCloudPlatform/anthos-security-blueprints/blob/master/restricting-traffic/kube-system/allow-kubedns-egress.yaml
var AnthosAllowKubeDNSIngressNetworkPolicy = &networkingv1.NetworkPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "allow-kube-dns-egress",
		Namespace: "kube-system",
		//Annotations: map[string]string{},
		//#configmanagement.gke.io/cluster-selector: ${CLUSTER_SELECTOR}
	},
	Spec: networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"k8s-app": "kube-dns",
			},
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				Ports: []networkingv1.NetworkPolicyPort{
					{Protocol: &tcp, Port: &port53},
					{Protocol: &udp, Port: &port53},
				},
				From: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector:       &metav1.LabelSelector{},
						NamespaceSelector: &metav1.LabelSelector{},
					},
				},
			},
		},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
	},
}

//// See: https://github.com/GoogleCloudPlatform/anthos-security-blueprints/blob/master/restricting-traffic/kube-system/allow-kubedns-ingress.yaml
//var AnthosAllowKubeDNSIngress = &Policy{
//	ObjectMeta: metav1.ObjectMeta{
//		Name:      "allow-kube-dns-ingress",
//		Namespace: "kube-system",
//		//Annotations:                nil,
//		//#configmanagement.gke.io/cluster-selector: ${CLUSTER_SELECTOR}
//	},
//	Spec: PolicySpec{
//		Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
//		TrafficMatcher: NewAll(
//			DestNamespaceMatcher("kube-system"),
//			&LabelMatcher{
//				Selector: NewKeyPathSelector(DestSelector, InternalSelector, NamespaceLabelsSelector),
//				Key:      "k8s-app",
//				Value:    "kube-dns",
//			},
//			NewAny(ProtocolMatcher(v1.ProtocolTCP), ProtocolMatcher(v1.ProtocolUDP)),
//			NumberedPortMatcher(53),
//			SourceIsInternalMatcher,
//		),
//		Directive: DirectiveAllow,
//	},
//}

/*
kind: NetworkPolicy2
apiVersion: networking.k8s.io/v1-alpha
metadata:
  name: allow-kube-dns-ingress
spec:
  trafficMatcher:
    - type: all
      matchers:
        - type: equal
          matchers:
          - type: keypath
            path: ["destination", "internal", "namespace"]
          - type: constant
            value: kube-system
        - type: any
          - matchers:
            - type: protocol
              value: tcp
            - type: protocol
              value: udp
        - type: port
          value: 53
        - type: not
          matcher:
            - type: equal
              matchers:
              - type: keypath
                path: ["source", "internal"]
              - type: constant
                value: nil
  directive: allow
*/

func DenyEgressFromNamespace(ns string) *Policy {
	return &Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("deny-egress-from-ns-%s", ns),
		},
		Spec: PolicySpec{
			Compatibility: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Priority:      10,
			TrafficMatcher: &TrafficEdge{
				Type: TrafficMatchTypeAll,
				Source: &PeerMatcher{
					Internal: &InternalPeerMatcher{
						Namespace: &StringMatcher{Value: ns},
					},
				},
			},
			Directive: DirectiveDeny,
		},
	}
}
