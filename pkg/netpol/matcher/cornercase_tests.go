package matcher

import (
	"github.com/mattfenwick/kube-prototypes/pkg/kube/netpol/examples"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var anySourceDestAndPort = &PeerPortMatcher{
	Peer: &AnywherePeerMatcher{},
	Port: &AllPortsAllProtocols{},
}

var anyTrafficPeer = &IngressEgressMatcher{Matchers: []*PeerPortMatcher{anySourceDestAndPort}}

func RunCornerCaseTests() {
	Describe("Allow none -- nil egress/ingress", func() {
		It("allow-no-ingress", func() {
			target := BuildTarget(examples.AllowNoIngress)

			Expect(target.Ingress).To(Equal(&IngressEgressMatcher{}))
			//Expect(target.Ingress).To(Equal(&IngressEgressMatcher{Matchers: []*PeerPortMatcher{}}))
			Expect(target.Egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			target := BuildTarget(examples.AllowNoEgress)

			Expect(target.Egress).To(Equal(&IngressEgressMatcher{}))
			Expect(target.Ingress).To(BeNil())
		})

		It("allow-neither", func() {
			target := BuildTarget(examples.AllowNoIngressAllowNoEgress)

			Expect(target.Ingress).To(Equal(&IngressEgressMatcher{}))
			Expect(target.Egress).To(Equal(&IngressEgressMatcher{}))
		})
	})

	Describe("Allow none -- empty ingress/egress", func() {
		It("allow-no-ingress", func() {
			target := BuildTarget(examples.AllowNoIngress_EmptyIngress)

			Expect(target.Ingress).To(Equal(&IngressEgressMatcher{}))
			Expect(target.Egress).To(BeNil())
		})

		It("allow-no-egress", func() {
			target := BuildTarget(examples.AllowNoEgress_EmptyEgress)

			Expect(target.Egress).To(Equal(&IngressEgressMatcher{}))
			Expect(target.Ingress).To(BeNil())
		})

		It("allow-neither", func() {
			target := BuildTarget(examples.AllowNoIngressAllowNoEgress_EmptyEgressEmptyIngress)

			Expect(target.Ingress).To(Equal(&IngressEgressMatcher{}))
			Expect(target.Egress).To(Equal(&IngressEgressMatcher{}))
		})
	})

	Describe("Allow all", func() {
		It("allow-all-ingress", func() {
			target := BuildTarget(examples.AllowAllIngress)

			Expect(target.Egress).To(BeNil())
			Expect(target.Ingress).To(Equal(anyTrafficPeer))
		})

		It("allow-all-egress", func() {
			target := BuildTarget(examples.AllowAllEgress)

			Expect(target.Egress).To(Equal(anyTrafficPeer))
			Expect(target.Ingress).To(BeNil())
		})

		It("allow-all-both", func() {
			target := BuildTarget(examples.AllowAllIngressAllowAllEgress)

			Expect(target.Egress).To(Equal(anyTrafficPeer))
			Expect(target.Ingress).To(Equal(anyTrafficPeer))
		})
	})

	Describe("Source/destination from slice of NetworkPolicyPeer", func() {
		It("allows all source/destination from an empty slice", func() {
			sds := BuildPeerMatchers("abc", []networkingv1.NetworkPolicyPeer{})
			Expect(sds).To(Equal([]PeerMatcher{&AnywherePeerMatcher{}}))
		})
	})

	Describe("Source/destination from NetworkPolicyPeer", func() {
		It("allow all pods in policy namespace", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer)
			Expect(sd).To(Equal(&AllPodsInPolicyNamespacePeerMatcher{Namespace: examples.Namespace}))
		})

		It("allow all pods in all namespaces", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer)
			Expect(sd).To(Equal(&AllPodsAllNamespacesPeerMatcher{}))
		})

		It("allow all pods in matching namespace", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer)
			Expect(sd).To(Equal(&AllPodsInMatchingNamespacesPeerMatcher{NamespaceSelector: *examples.SelectorAB}))
		})

		It("allow all pods in policy namespace -- empty pod selector", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInPolicyNamespacePeer_EmptyPodSelector)
			Expect(sd).To(Equal(&AllPodsInPolicyNamespacePeerMatcher{Namespace: examples.Namespace}))
		})

		It("allow all pods in all namespaces -- empty pod selector", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInAllNamespacesPeer_EmptyPodSelector)
			Expect(sd).To(Equal(&AllPodsAllNamespacesPeerMatcher{}))
		})

		It("allow all pods in matching namespace -- empty pod selector", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowAllPodsInMatchingNamespacesPeer_EmptyPodSelector)
			Expect(sd).To(Equal(&AllPodsInMatchingNamespacesPeerMatcher{NamespaceSelector: *examples.SelectorAB}))
		})

		It("allow matching pods in policy namespace", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInPolicyNamespacePeer)
			Expect(sd).To(Equal(&MatchingPodsInPolicyNamespacePeerMatcher{PodSelector: *examples.SelectorCD, Namespace: examples.Namespace}))
		})

		It("allow matching pods in all namespaces", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInAllNamespacesPeer)
			Expect(sd).To(Equal(&MatchingPodsInAllNamespacesPeerMatcher{PodSelector: *examples.SelectorEF}))
		})

		It("allow matching pods in matching namespace", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowMatchingPodsInMatchingNamespacesPeer)
			Expect(sd).To(Equal(&MatchingPodsInMatchingNamespacesPeerMatcher{
				PodSelector:       *examples.SelectorGH,
				NamespaceSelector: *examples.SelectorAB,
			}))
		})

		It("allow ipblock", func() {
			sd := BuildPeerMatcher(examples.Namespace, examples.AllowIPBlockPeer)
			Expect(sd).To(Equal(&IPBlockPeerMatcher{
				&networkingv1.IPBlock{CIDR: "10.0.0.1/24",
					Except: []string{
						"10.0.0.2",
					},
				},
			}))
		})
	})

	Describe("Port from slice of NetworkPolicyPort", func() {
		It("allows all ports and all protocols from an empty slice", func() {
			sds := BuildPortMatchers([]networkingv1.NetworkPolicyPort{})
			Expect(sds).To(Equal([]PortMatcher{&AllPortsAllProtocols{}}))
		})
	})

	Describe("Port from NetworkPolicyPort", func() {
		It("allow all ports on protocol", func() {
			sd := BuildPortMatcher(examples.AllowAllPortsOnProtocol)
			Expect(sd).To(Equal(&AllPortsOnProtocol{Protocol: v1.ProtocolSCTP}))
		})

		It("allow numbered port on protocol", func() {
			sd := BuildPortMatcher(examples.AllowNumberedPortOnProtocol)
			Expect(sd).To(Equal(&ExactPortProtocol{
				Protocol: v1.ProtocolTCP,
				Port:     intstr.FromInt(9001),
			}))
		})

		It("allow named port on protocol", func() {
			sd := BuildPortMatcher(examples.AllowNamedPortOnProtocol)
			Expect(sd).To(Equal(&ExactPortProtocol{
				Protocol: v1.ProtocolUDP,
				Port:     intstr.FromString("hello"),
			}))
		})
	})
}
