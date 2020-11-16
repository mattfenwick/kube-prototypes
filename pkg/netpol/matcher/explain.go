package matcher

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func Explain(policies *Policy) string {
	var lines []string
	for _, t := range policies.Targets {
		lines = append(lines, t.GetPrimaryKey())
		if t.Ingress != nil {
			lines = append(lines, "  ingress:")
			lines = append(lines, ExplainTrafficPeers(t.Ingress)...)
		}
		if t.Egress != nil {
			lines = append(lines, "  egress:")
			lines = append(lines, ExplainTrafficPeers(t.Egress)...)
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func ExplainTrafficPeers(tp *IngressEgressMatcher) []string {
	var lines []string
	for _, sd := range tp.Matchers {
		var sourceDest, port string
		switch t := sd.Peer.(type) {
		case *MatchingPodsInAllNamespacesPeerMatcher:
			sourceDest = fmt.Sprintf("pods matching %s in all namespaces",
				SerializeLabelSelector(t.PodSelector))
		case *MatchingPodsInMatchingNamespacesPeerMatcher:
			sourceDest = fmt.Sprintf("pods matching %s in namespaces matching %s",
				SerializeLabelSelector(t.PodSelector),
				SerializeLabelSelector(t.NamespaceSelector))
		case *AllPodsInMatchingNamespacesPeerMatcher:
			sourceDest = fmt.Sprintf("all pods in namespaces matching %s",
				SerializeLabelSelector(t.NamespaceSelector))
		case *AllPodsInPolicyNamespacePeerMatcher:
			sourceDest = fmt.Sprintf("all pods in namespace %s", t.Namespace)
		case *MatchingPodsInPolicyNamespacePeerMatcher:
			sourceDest = fmt.Sprintf("pods matching %s in namespace %s",
				SerializeLabelSelector(t.PodSelector), t.Namespace)
		case *AllPodsAllNamespacesPeerMatcher:
			sourceDest = "all pods in all namespaces"
		case *AnywherePeerMatcher:
			sourceDest = "anywhere: all pods in all namespaces and all IPs"
		case *IPBlockPeerMatcher:
			sourceDest = fmt.Sprintf("IPBlock: cidr %s, except %+v", t.IPBlock.CIDR, t.IPBlock.Except)
		default:
			panic(errors.Errorf("unexpected PeerMatcher type %T", t))
		}
		switch p := sd.Port.(type) {
		case *AllPortsOnProtocolMatcher:
			port = fmt.Sprintf("all ports on protocol %s", p.Protocol)
		case *AllPortsAllProtocolsMatcher:
			port = "all ports all protocols"
		case *ExactPortProtocolMatcher:
			port = fmt.Sprintf("port %s on protocol %s", p.Port.String(), p.Protocol)
		default:
			panic(errors.Errorf("unexpected Port type %T", p))
		}
		lines = append(lines, "  - "+sourceDest, "    "+port)
	}

	return lines
}
