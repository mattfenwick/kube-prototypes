package matcher

import (
	"encoding/json"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/kube"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceDestAndPort struct {
	SourceDest SourceDest
	Port       Port
}

func (sdap *SourceDestAndPort) Allows(rt *ResolvedTraffic) bool {
	if !sdap.Port.Allows(rt.Traffic.Port) {
		return false
	}
	return sdap.SourceDest.Allows(rt.Traffic)
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
// 8. everything
//   - don't have anything at all -- i.e. empty []NetworkPolicyPeer
//

type SourceDest interface {
	Allows(t *Traffic) bool
}

// AllPodsInPolicyNamespaceSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is empty or nil
// - NamespaceSelector is nil
// - IPBlock is nil
type AllPodsInPolicyNamespaceSourceDest struct {
	Namespace string
}

func (p *AllPodsInPolicyNamespaceSourceDest) Allows(t *Traffic) bool {
	if t.IsExternal {
		return false
	}
	return t.Namespace == p.Namespace
}

func (p *AllPodsInPolicyNamespaceSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":      "all pods in policy namespace",
		"Namespace": p.Namespace,
	})
}

// AllPodsAllNamespacesSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is nil or empty
// - NamespaceSelector is empty (but not nil!)
// - IPBlock is nil
type AllPodsAllNamespacesSourceDest struct{}

func (a *AllPodsAllNamespacesSourceDest) Allows(t *Traffic) bool {
	return !t.IsExternal
}

func (a *AllPodsAllNamespacesSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all pods in all namespaces",
	})
}

// AllPodsInMatchingNamespacesSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is nil or empty
// - NamespaceSelector is not empty
// - IPBlock is nil
type AllPodsInMatchingNamespacesSourceDest struct {
	NamespaceSelector metav1.LabelSelector
}

func (a *AllPodsInMatchingNamespacesSourceDest) Allows(t *Traffic) bool {
	if t.IsExternal {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(t.NamespaceLabels, a.NamespaceSelector)
}

func (a *AllPodsInMatchingNamespacesSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":              "all pods in matching namespaces",
		"NamespaceSelector": a.NamespaceSelector,
	})
}

// MatchingPodsInPolicyNamespaceSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is not empty
// - NamespaceSelector is nil
// - IPBlock is nil
type MatchingPodsInPolicyNamespaceSourceDest struct {
	PodSelector metav1.LabelSelector
	Namespace   string
}

func (p *MatchingPodsInPolicyNamespaceSourceDest) Allows(t *Traffic) bool {
	if t.IsExternal {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(t.PodLabels, p.PodSelector) && t.Namespace == p.Namespace
}

func (p *MatchingPodsInPolicyNamespaceSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":        "matchings pods in policy namespace",
		"PodSelector": p.PodSelector,
		"Namespace":   p.Namespace,
	})
}

// MatchingPodsInAllNamespacesSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is not nil
// - NamespaceSelector is empty (but not nil!)
// - IPBlock is nil
type MatchingPodsInAllNamespacesSourceDest struct {
	PodSelector metav1.LabelSelector
}

func (p *MatchingPodsInAllNamespacesSourceDest) Allows(t *Traffic) bool {
	if t.IsExternal {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(t.PodLabels, p.PodSelector)
}

func (p *MatchingPodsInAllNamespacesSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":        "pods in all namespaces",
		"PodSelector": p.PodSelector,
	})
}

// MatchingPodsInMatchingNamespacesSourceDest models the case where in NetworkPolicyPeer:
// - PodSelector is not nil
// - NamespaceSelector is not empty
// - IPBlock is nil
type MatchingPodsInMatchingNamespacesSourceDest struct {
	PodSelector       metav1.LabelSelector
	NamespaceSelector metav1.LabelSelector
}

func (s *MatchingPodsInMatchingNamespacesSourceDest) Allows(t *Traffic) bool {
	if t.IsExternal {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(t.NamespaceLabels, s.NamespaceSelector) &&
		kube.IsLabelsMatchLabelSelector(t.PodLabels, s.PodSelector)
}

func (s *MatchingPodsInMatchingNamespacesSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":              "matching pods in matching namespaces",
		"PodSelector":       s.PodSelector,
		"NamespaceSelector": s.NamespaceSelector,
	})
}

// AnywhereSourceDest models the case where NetworkPolicy(E|In)gressRule.(From|To) is empty
type AnywhereSourceDest struct{}

func (a *AnywhereSourceDest) Allows(t *Traffic) bool {
	return true
}

func (a *AnywhereSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "anywhere",
	})
}

// IPBlockSourceDest models the case where IPBlock is not nil, and both
// PodSelector and NamespaceSelector are nil
type IPBlockSourceDest struct {
	// required
	IPBlock *v1.IPBlock
}

func (ibsd *IPBlockSourceDest) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":   "IPBlock",
		"CIDR":   ibsd.IPBlock.CIDR,
		"Except": ibsd.IPBlock.Except,
	})
}

func (ibsd *IPBlockSourceDest) Allows(t *Traffic) bool {
	return kube.IsIPBlockMatchForIP(t.IP, ibsd.IPBlock)
}
