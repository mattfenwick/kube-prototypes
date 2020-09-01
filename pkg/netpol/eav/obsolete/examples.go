package obsolete

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// All/All
var AllTargetsAllPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnythingPeerMatcher{},
	PeerMatcher:   &AnythingPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// All/Internal
var AllTargetsInternalPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnythingPeerMatcher{},
	PeerMatcher:   &AnyInternalPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// All/External
var AllTargetsExternalPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnythingPeerMatcher{},
	PeerMatcher:   &AnyExternalPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// All/None
var AllTargetsNoPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnythingPeerMatcher{},
	PeerMatcher:   &NothingPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// Internal/All
var InternalTargetsAllPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnyInternalPeerMatcher{},
	PeerMatcher:   &AnythingPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// Internal/Internal
var InternalTargetsInternalPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnyInternalPeerMatcher{},
	PeerMatcher:   &AnyInternalPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// Internal/External
var InternalTargetsExternalPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnyInternalPeerMatcher{},
	PeerMatcher:   &AnyExternalPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// Internal/None
var InternalTargetsNoPeers = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnyInternalPeerMatcher{},
	PeerMatcher:   &NothingPeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}

// TODO these are probably all useless -- since they don't match any targets:
// External/All
// External/Internal
// External/External
// External/None
// None/All
// None/Internal
// None/External
// None/None

var PodLabelTargetNamespaceLabelPeer = &Policy{
	Type: networkingv1.PolicyTypeIngress,
	TargetMatcher: &InternalLabelsPeerMatcher{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "web",
			},
		},
	},
	PeerMatcher: &InternalLabelsPeerMatcher{
		NamespaceSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"stage": "dev",
			},
		},
	},
	PortMatcher: AnyProtocolPortMatcher,
}

var SameNamespaceTargetAndPeer = &Policy{
	Type:          networkingv1.PolicyTypeIngress,
	TargetMatcher: &AnyInternalPeerMatcher{},
	PeerMatcher:   &SameNamespacePeerMatcher{},
	PortMatcher:   AnyProtocolPortMatcher,
}
