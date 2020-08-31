package eav

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Traffic struct {
	Source      *Peer
	Destination *Peer

	Protocol v1.Protocol
	Port     intstr.IntOrString
}

type Peer struct {
	Internal *struct {
		PodLabels       map[string]string
		PodName         string
		NamespaceLabels map[string]string
		Namespace       string
		NodeLabels      map[string]string
		NodeName        string
	}
	IP string
}

func (tc *Peer) IsExternal() bool {
	return tc.Internal == nil
}
