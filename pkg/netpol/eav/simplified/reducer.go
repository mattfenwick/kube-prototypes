package eav

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	//return &networkingv1.NetworkPolicy{
	//	ObjectMeta: v1.ObjectMeta{
	//		Name: np.Name,
	//	},
	//	Spec: networkingv1.NetworkPolicySpec{
	//		PodSelector: v1.LabelSelector{},
	//		Ingress:     nil,
	//		Egress:      nil,
	//		PolicyTypes: nil,
	//	},
	//}
	return []*networkingv1.NetworkPolicy{}
}

func Flatten(matcher TrafficMatcher) {
	switch t := matcher.(type) {
	case *Equal:
		var s, d []*
	case *All:
	case *Any:
	case *Not:
	case *Bool:
	case *InArray:
	}
}
