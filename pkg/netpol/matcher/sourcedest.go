package matcher

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceDestAndPort struct {
	SourceDest SourceDest
	Port       Port
}

func (sdap *SourceDestAndPort) Allows(td *TrafficDirection) bool {
	if !sdap.Port.Allows(td.Port) {
		return false
	}
	return sdap.SourceDest.Allows(td.SourceDest)
}

// NetworkPolicyPeer possibilities:
// 1. PodSelector:
//   - empty/nil
//   - not empty
// 2. NamespaceSelector
//   - nil
//   - empty
//   - not empty
// 3. IPBlock
//   - nil
//   - not nil
//
// Combined:
// 1. all pods in policy namespace
//   - empty/nil PodSelector
//   - nil NamespaceSelector
//
// 2. all pods in all namespaces
//   - empty/nil PodSelector
//   - empty NamespaceSelector
//
// 3. all pods in matching namespaces
//   - empty/nil PodSelector
//   - not empty NamespaceSelector
//
// 4. matching pods in policy namespace
//   - not empty PodSelector
//   - nil NamespaceSelector
//
// 5. matching pods in all namespaces
//   - not empty PodSelector
//   - empty NamespaceSelector
//
// 6. matching pods in matching namespaces
//   - not empty PodSelector
//   - not empty NamespaceSelector
//
// 7. matching IPBlock
//   - IPBlock
//
// 8. ??? is this possible ???
//   - everything nil
//

type SourceDest interface {
	Allows(t TrafficSourceDest) bool
}

// PodsInAllNamespacesSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is not nil
// - NamespaceSelector is empty (but not nil!)
// - IPBlock is nil
type PodsInAllNamespacesSourceDest struct {
	PodSelector metav1.LabelSelector
}

func (p *PodsInAllNamespacesSourceDest) Allows(t TrafficSourceDest) bool {
	if t.IsExternal() {
		return false
	}
	return isLabelsMatchLabelSelector(t.GetPodLabels(), p.PodSelector)
}

// SpecificPodsNamespaceSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is not nil
// - NamespaceSelector is not empty
// - IPBlock is nil
type SpecificPodsNamespaceSourceDest struct {
	PodSelector       metav1.LabelSelector
	NamespaceSelector metav1.LabelSelector
}

func (s *SpecificPodsNamespaceSourceDest) Allows(t TrafficSourceDest) bool {
	if t.IsExternal() {
		return false
	}
	return isLabelsMatchLabelSelector(t.GetNamespaceLabels(), s.NamespaceSelector) &&
		isLabelsMatchLabelSelector(t.GetPodLabels(), s.PodSelector)
}

// AllPodsInNamespaceSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is nil or empty
// - NamespaceSelector is not empty
// - IPBlock is nil
type AllPodsInNamespaceSourceDest struct {
	NamespaceSelector metav1.LabelSelector
}

func (a *AllPodsInNamespaceSourceDest) Allows(t TrafficSourceDest) bool {
	if t.IsExternal() {
		return false
	}
	return isLabelsMatchLabelSelector(t.GetNamespaceLabels(), a.NamespaceSelector)
}

// AllPodsInPolicyNamespaceSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is empty or nil
// - NamespaceSelector is nil
// - IPBlock is nil
type AllPodsInPolicyNamespaceSourceDest struct {
	Namespace string
}

func (p *AllPodsInPolicyNamespaceSourceDest) Allows(t TrafficSourceDest) bool {
	if t.IsExternal() {
		return false
	}
	return t.GetNamespace() == p.Namespace
}

// PodsInPolicyNamespaceSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is not empty
// - NamespaceSelector is nil
// - IPBlock is nil
type PodsInPolicyNamespaceSourceDest struct {
	PodSelector metav1.LabelSelector
	Namespace   string
}

func (p *PodsInPolicyNamespaceSourceDest) Allows(t TrafficSourceDest) bool {
	if t.IsExternal() {
		return false
	}
	return isLabelsMatchLabelSelector(t.GetPodLabels(), p.PodSelector) && t.GetNamespace() == p.Namespace
}

// AllPodsAllNamespacesSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is nil or empty
// - NamespaceSelector is empty (but not nil!)
// - IPBlock is nil
type AllPodsAllNamespacesSourceDest struct{}

func (a *AllPodsAllNamespacesSourceDest) Allows(t TrafficSourceDest) bool {
	return !t.IsExternal()
}

// AnywhereSourceDest models the case where NetworkPolicy(E|In)gressRule.(From|To) is empty
type AnywhereSourceDest struct{}

func (a *AnywhereSourceDest) Allows(t TrafficSourceDest) bool {
	return true
}

// IPBlockSourceDest models the case where IPBlock is not nil, and both
// PodSelector and NamespaceSelector are nil
type IPBlockSourceDest struct {
	CIDR   string
	Except []string
}

func (a *IPBlockSourceDest) Allows(t TrafficSourceDest) bool {
	//ip, ipnet, err := net.ParseCIDR(t.GetIP())
	//if err != nil {
	//	panic(err)
	//}
	//panic(errors.Errorf("TODO -- ip %+v, ipnet %+v", ip, ipnet))
	panic("TODO")
}

func isLabelsMatchLabelSelector(labels map[string]string, labelSelector metav1.LabelSelector) bool {
	for key, val := range labelSelector.MatchLabels {
		if labels[key] == val {
			return true
		}
	}
	for _, exp := range labelSelector.MatchExpressions {
		switch exp.Operator {
		case metav1.LabelSelectorOpIn:
			val, ok := labels[exp.Key]
			if !ok {
				return false
			}
			for _, v := range exp.Values {
				if v == val {
					return true
				}
			}
			return false
		case metav1.LabelSelectorOpNotIn:
			val, ok := labels[exp.Key]
			if !ok {
				// see https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements
				//   even for NotIn -- if the key isn't there, it's not a match
				return false
			}
			for _, v := range exp.Values {
				if v == val {
					return false
				}
			}
			return true
		case metav1.LabelSelectorOpExists:
			_, ok := labels[exp.Key]
			return !ok
		case metav1.LabelSelectorOpDoesNotExist:
			_, ok := labels[exp.Key]
			return !ok
		default:
			panic("invalid operator")
		}
	}
	return false
}
