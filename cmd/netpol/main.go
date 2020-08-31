package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/examples"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/kube"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/matcher"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	//v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func main() {
	// 1. probe all pods in a namespace
	//  - on a few different ports ?  or on the port opened for the services?  <- get the right ports
	// 2. apply a blanket-deny netpol
	//  - probe pods again
	// 3. start opening communication between pods
	// 4. figure out some visualization of connectivity

	//builder := netpol.NetworkPolicySpecBuilder{
	//	Spec:      v1.NetworkPolicySpec{},
	//	Name:      "",
	//	Namespace: "",
	//}

	k8s, err := kube.NewKubernetes()
	doOrDie(err)

	if true {
		err = k8s.CleanNetworkPolicies("default")
		doOrDie(err)

		var allCreated []*networkingv1.NetworkPolicy
		for _, np := range examples.AllExamples {
			createdNp, err := k8s.CreateNetworkPolicy(np)
			allCreated = append(allCreated, createdNp)
			doOrDie(err)
			//explanation := netpol.ExplainPolicy(np)
			explanation := netpol.ExplainPolicy(createdNp)
			fmt.Printf("policy explanation for %s:\n%s\n\n", np.Name, explanation.PrettyPrint())

			matcherExplanation := matcher.Explain(matcher.BuildNetworkPolicy(createdNp))
			fmt.Printf("\nmatcher explanation: %s\n\n", matcherExplanation)

			reduced := netpol.Reduce(createdNp)
			fmt.Println(netpol.NodePrettyPrint(reduced))
			fmt.Println()

			createdNpBytes, err := json.MarshalIndent(createdNp, "", "  ")
			doOrDie(err)
			fmt.Printf("created netpol:\n\n%s\n\n", createdNpBytes)

			matcherPolicy := matcher.BuildNetworkPolicy(createdNp)
			matcherPolicyBytes, err := json.MarshalIndent(matcherPolicy, "", "  ")
			doOrDie(err)
			fmt.Printf("created matcher netpol:\n\n%s\n\n", matcherPolicyBytes)
			isAllowed, allowers, matchingTargets := matcherPolicy.IsTrafficAllowed(&matcher.ResolvedTraffic{
				Traffic: matcher.NewPodTraffic(
					map[string]string{
						"app": "bookstore",
					},
					map[string]string{},
					"not-default",
					true,
					&matcher.PortProtocol{
						Protocol: v1.ProtocolTCP,
						Port:     intstr.FromInt(9800),
					},
					"1.2.3.4"),
				Target: &matcher.ResolvedPodTarget{
					PodLabels: map[string]string{
						"app": "web",
					},
					NamespaceLabels: nil,
					Namespace:       "default",
				},
			})
			fmt.Printf("is allowed?  %t\n - allowers: %+v\n - matching targets: %+v\n", isAllowed, allowers, matchingTargets)
		}

		netpols := matcher.BuildNetworkPolicies(allCreated)
		bytes, err := json.MarshalIndent(netpols, "", "  ")
		doOrDie(err)
		fmt.Printf("full network policies:\n\n%s\n\n", bytes)
		fmt.Printf("\nexplained:\n%s\n", matcher.Explain(netpols))

		netpolsExamples := matcher.BuildNetworkPolicy(examples.ExampleComplicatedNetworkPolicy())
		fmt.Printf("complicated example explained:\n%s\n", matcher.Explain(netpolsExamples))
	}

	if false {
		ns := os.Args[1]
		// TODO add another pod in a different namespace to illustrate cross-namespace behavior
		if false {
			probeContainerToContainer(ns, k8s, 8443)
		}

		probeContainerToService(ns, k8s)
	}
}

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func probeContainerToContainer(ns string, k8s *kube.Kubernetes, port int) {
	pods, err := k8s.ClientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	doOrDie(err)
	var items []string
	for _, p := range pods.Items {
		// the order of status containers doesn't necessarily match the order of spec containers for the same
		//   pod, so let's just always use the spec to be consistent and reduce confusion!
		for _, c := range p.Spec.Containers {
			items = append(items, c.Name)
		}
	}
	table := netpol.NewStringTruthTable(items)
	for _, fromPod := range pods.Items {
		for _, fromCont := range fromPod.Spec.Containers {
			fromContainer := fromCont.Name
			for _, toPod := range pods.Items {
				for _, toCont := range toPod.Spec.Containers {
					log.Infof("Probing in ns %s: %s, %s", ns, fromPod.Name, toPod.Name)
					//connected, err := k8s.ProbeWithPod(fromPod, toPod, port)
					connected, curlExitCode, err := k8s.ProbeFromContainerToPod(&kube.ProbeFromContainerToPod{
						FromNamespace:      fromPod.Namespace,
						FromPod:            fromPod.Name,
						FromContainer:      fromContainer,
						ToIP:               toPod.Status.PodIP,
						ToPort:             int(toCont.Ports[0].ContainerPort),
						ToNamespace:        toPod.Namespace,
						ToPod:              toPod.Name,
						CurlTimeoutSeconds: 5,
					})
					log.Warningf("curl exit code: %d", curlExitCode)
					if err != nil {
						log.Errorf("unable to make main observation on %s -> %s: %+v", fromPod.Name, toPod.Name, err)
					}
					if !connected {
						log.Warnf("FAILED CONNECTION FOR WHITELISTED PODS %s -> %s !!!! ", fromPod.Name, toPod.Name)
					}
					table.Set(fromContainer, toCont.Name, fmt.Sprintf("%d", curlExitCode))
				}
			}
		}
	}

	table.Table().Render()
}

func probeContainerToService(namespace string, k8s *kube.Kubernetes) {
	pods, err := k8s.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	doOrDie(err)

	services, err := k8s.ClientSet.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	doOrDie(err)

	var froms []string
	for _, p := range pods.Items {
		// the order of status containers doesn't necessarily match the order of spec containers for the same
		//   pod, so let's just always use the spec to be consistent and reduce confusion!
		for _, c := range p.Spec.Containers {
			froms = append(froms, c.Name)
		}
	}
	var tos []string
	for _, s := range services.Items {
		tos = append(tos, s.Name)
	}
	table := netpol.NewStringTruthTableWithFromsTo(froms, tos)

	for _, fromPod := range pods.Items {
		for _, fromCont := range fromPod.Spec.Containers {
			fromContainer := fromCont.Name
			for _, toService := range services.Items {
				log.Infof("Probing in ns %s: %s, %s", namespace, fromPod.Name, toService.Name)
				//connected, err := k8s.ProbeWithPod(fromPod, toPod, port)
				connected, curlExitCode, err := k8s.ProbeFromContainerToPod(&kube.ProbeFromContainerToPod{
					FromNamespace:      fromPod.Namespace,
					FromPod:            fromPod.Name,
					FromContainer:      fromContainer,
					ToIP:               toService.Name,
					ToPort:             int(toService.Spec.Ports[0].Port),
					ToNamespace:        namespace,
					ToPod:              "(actually a service)",
					CurlTimeoutSeconds: 5,
				})
				log.Warningf("curl exit code: %d", curlExitCode)
				if err != nil {
					log.Errorf("unable to make main observation on %s -> %s: %+v", fromPod.Name, toService.Name, err)
				}
				if !connected {
					log.Warnf("FAILED CONNECTION FOR WHITELISTED PODS %s -> %s !!!! ", fromPod.Name, toService.Name)
				}
				table.Set(fromContainer, toService.Name, fmt.Sprintf("%d", curlExitCode))
			}
		}
	}

	table.Table().Render()
}
