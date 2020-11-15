package obsolete

import (
	"github.com/mattfenwick/kube-prototypes/pkg/kube"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PeerMatcher interface {
	IsPeerMatch(peer *Peer) bool
}

// NothingPeerMatcher matches nothing
type NothingPeerMatcher struct{}

func (npm *NothingPeerMatcher) IsPeerMatch(peer *Peer) bool {
	return false
}

// AnythingPeerMatcher matches anything
type AnythingPeerMatcher struct{}

func (apm *AnythingPeerMatcher) IsPeerMatch(peer *Peer) bool {
	return true
}

// AnyInternalPeerMatcher matches any pod in the same kube clusters
type AnyInternalPeerMatcher struct{}

func (aipm *AnyInternalPeerMatcher) IsPeerMatch(peer *Peer) bool {
	return peer.Internal != nil
}

// AnyExternalPeerMatcher matches anything NOT in the same kube cluster
type AnyExternalPeerMatcher struct{}

func (aepm *AnyExternalPeerMatcher) IsPeerMatch(peer *Peer) bool {
	return peer.Internal == nil
}

// InternalLabelsPeerMatcher matches pods based on:
// - namespace name and labels
// - node name and labels
// - pod name and labels
// It never matches external Peers.
// Corner cases:
// - all selectors empty: matches all internal traffic (should use AnyInternalPeerMatcher instead)
// - must match all selectors -- Namespace, Node, and Pod
//   - if it matches Namespace but not Node -- no match
//   - use case: match by Namespace but ignore Nodes and Pods: specify a non-empty Namespace selector,
//     and empty Node and Pod selectors
type InternalLabelsPeerMatcher struct {
	Namespace         string
	NamespaceSelector metav1.LabelSelector
	Node              string
	NodeSelector      metav1.LabelSelector
	Pod               string
	PodSelector       metav1.LabelSelector
}

func (ilpm *InternalLabelsPeerMatcher) IsPeerMatch(peer *Peer) bool {
	if peer.Internal == nil {
		return false
	}
	return kube.IsLabelsMatchLabelSelector(peer.Internal.NamespaceLabels, ilpm.NamespaceSelector) &&
		kube.IsLabelsMatchLabelSelector(peer.Internal.NodeLabels, ilpm.NodeSelector) &&
		kube.IsLabelsMatchLabelSelector(peer.Internal.PodLabels, ilpm.PodSelector) &&
		kube.IsNameMatch(peer.Internal.Namespace, ilpm.Namespace) &&
		kube.IsNameMatch(peer.Internal.Node, ilpm.Node) &&
		kube.IsNameMatch(peer.Internal.Pod, ilpm.Pod)
}

type SameNamespacePeerMatcher struct{}

func (snpm *SameNamespacePeerMatcher) IsPeerMatch(peer *Peer) bool {
	if peer.Internal == nil {
		return false
	}
	panic("TODO -- return peer.Internal.Namespace == target.Namespace")
}

// IPBlockPeerMatcher matches based on the peer IP
type IPBlockPeerMatcher struct {
	IPBlock *v1.IPBlock
}

func (ibpm *IPBlockPeerMatcher) IsPeerMatch(peer *Peer) bool {
	return kube.IsIPBlockMatchForIP(peer.IP, ibpm.IPBlock)
}
