package matcher

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func Explain(policies *NetworkPolicies) string {
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

func ExplainTrafficPeers(tp *TrafficPeers) []string {
	var lines []string
	for _, sd := range tp.SourcesOrDests {
		var sourceDest, port string
		switch t := sd.SourceDest.(type) {
		case *MatchingPodsInAllNamespacesSourceDest:
			sourceDest = fmt.Sprintf("pods matching %s in all namespaces",
				SerializeLabelSelector(t.PodSelector))
		case *MatchingPodsInMatchingNamespacesSourceDest:
			sourceDest = fmt.Sprintf("pods matching %s in namespaces matching %s",
				SerializeLabelSelector(t.PodSelector),
				SerializeLabelSelector(t.NamespaceSelector))
		case *AllPodsInMatchingNamespacesSourceDest:
			sourceDest = fmt.Sprintf("all pods in namespaces matching %s",
				SerializeLabelSelector(t.NamespaceSelector))
		case *AllPodsInPolicyNamespaceSourceDest:
			sourceDest = fmt.Sprintf("all pods in namespace %s", t.Namespace)
		case *MatchingPodsInPolicyNamespaceSourceDest:
			sourceDest = fmt.Sprintf("pods matching %s in namespace %s",
				SerializeLabelSelector(t.PodSelector), t.Namespace)
		case *AllPodsAllNamespacesSourceDest:
			sourceDest = "all pods in all namespaces"
		case *AnywhereSourceDest:
			sourceDest = "anywhere: all pods in all namespaces and all IPs"
		case *IPBlockSourceDest:
			sourceDest = fmt.Sprintf("IPBlock: cidr %s, except %+v", t.IPBlock.CIDR, t.IPBlock.Except)
		default:
			panic(errors.Errorf("unexpected SourceDest type %T", t))
		}
		switch p := sd.Port.(type) {
		case *AllPortsOnProtocol:
			port = fmt.Sprintf("all ports on protocol %s", p.Protocol)
		case *AllPortsAllProtocols:
			port = "all ports all protocols"
		case *PortProtocol:
			port = fmt.Sprintf("port %s on protocol %s", p.Port.String(), p.Protocol)
		default:
			panic(errors.Errorf("unexpected Port type %T", p))
		}
		lines = append(lines, "  - "+sourceDest, "    "+port)
	}

	return lines
}
