package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/examples"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/kube"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/matcher"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	//v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpLogger(logLevelStr string) error {
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		return errors.Wrapf(err, "unable to parse the specified log level: '%s'", logLevel)
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Infof("log level set to '%s'", log.GetLevel())
	return nil
}

type Flags struct {
	Verbosity string
}

func setupCommand() *cobra.Command {
	flags := &Flags{}
	command := &cobra.Command{
		Use:   "kube-prototypes",
		Short: "kube hacking",
		Long:  "kube hacking",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return SetUpLogger(flags.Verbosity)
		},
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.AddCommand(SetupNetpolCommand())
	command.AddCommand(SetupProbeCommand())

	return command
}

func SetupNetpolCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "netpols",
		Short: "netpol hacking",
		Long:  "do stuff with network policies",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			mungeNetworkPolicies()
		},
	}

	return command
}

func SetupProbeCommand() *cobra.Command {
	var namespaces []string
	var probeContainers, probeServices bool

	command := &cobra.Command{
		Use:   "probe",
		Short: "probe connectivity",
		Long:  "probe pod -> pod and pod -> service connectivity",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runProbe(namespaces, probeContainers, probeServices)
		},
	}

	command.Flags().StringSliceVar(&namespaces, "nss", []string{}, "namespaces to probe")
	command.MarkFlagRequired("nss")

	command.Flags().BoolVar(&probeContainers, "cont", true, "probe containers")
	command.Flags().BoolVar(&probeServices, "svc", false, "probe services")

	return command
}

func main() {
	command := setupCommand()
	err := errors.Wrapf(command.Execute(), "run root command")
	doOrDie(err)

	// 1. probe all pods in a namespace
	//  - on a few different ports ?  or on the port opened for the services?  <- get the right ports
	// 2. apply a blanket-deny netpol
	//  - probe pods again
	// 3. start opening communication between pods
	// 4. figure out some visualization of connectivity
}

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func mungeNetworkPolicies() {
	k8s, err := kube.NewKubernetes()
	doOrDie(err)

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

func runProbe(namespaces []string, probeContainers bool, probeServices bool) {
	k8s, err := kube.NewKubernetes()
	doOrDie(err)

	if probeContainers {
		probeContainerToContainer(namespaces, k8s)
	}

	if probeServices {
		probeContainerToService(namespaces, k8s)
	}
}

func podKey(pod v1.Pod) string {
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}

func serviceKey(svc v1.Service) string {
	return fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
}

func getPods(k8s *kube.Kubernetes, namespaces []string) ([]v1.Pod, error) {
	var pods []v1.Pod
	for _, ns := range namespaces {
		podList, err := k8s.ClientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get pods in namespace %s", ns)
		}
		pods = append(pods, podList.Items...)
	}
	return pods, nil
}

func getServices(k8s *kube.Kubernetes, namespaces []string) ([]v1.Service, error) {
	var services []v1.Service
	for _, ns := range namespaces {
		serviceList, err := k8s.ClientSet.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get pods in namespace %s", ns)
		}
		services = append(services, serviceList.Items...)
	}
	return services, nil
}

func probeContainerToContainer(namespaces []string, k8s *kube.Kubernetes) {
	pods, err := getPods(k8s, namespaces)
	doOrDie(err)

	var items []string
	for _, p := range pods {
		items = append(items, podKey(p))
	}

	table := netpol.NewStringTruthTable(items)
	for _, fromPod := range pods {
		for _, fromCont := range fromPod.Spec.Containers {
			fromContainer := fromCont.Name
			for _, toPod := range pods {
				for _, toCont := range toPod.Spec.Containers {
					log.Infof("Probing %s -> %s", podKey(fromPod), podKey(toPod))
					//connected, err := k8s.ProbeWithPod(fromPod, toPod, port)
					if len(toCont.Ports) == 0 {
						table.Set(podKey(fromPod), podKey(toPod), "no ports")
						continue
					}
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
					table.Set(podKey(fromPod), podKey(toPod), fmt.Sprintf("%d", curlExitCode))
				}
			}
		}
	}

	table.Table().Render()
}

func probeContainerToService(namespaces []string, k8s *kube.Kubernetes) {
	pods, err := getPods(k8s, namespaces)
	doOrDie(err)

	services, err := getServices(k8s, namespaces)
	doOrDie(err)

	var froms []string
	for _, p := range pods {
		froms = append(froms, podKey(p))
	}
	var tos []string
	for _, s := range services {
		tos = append(tos, serviceKey(s))
	}
	table := netpol.NewStringTruthTableWithFromsTo(froms, tos)

	for _, fromPod := range pods {
		for _, fromCont := range fromPod.Spec.Containers {
			fromContainer := fromCont.Name
			for _, toService := range services {
				log.Infof("Probing %s -> %s", podKey(fromPod), serviceKey(toService))
				//connected, err := k8s.ProbeWithPod(fromPod, toPod, port)
				connected, curlExitCode, err := k8s.ProbeFromContainerToPod(&kube.ProbeFromContainerToPod{
					FromNamespace:      fromPod.Namespace,
					FromPod:            fromPod.Name,
					FromContainer:      fromContainer,
					ToIP:               toService.Name,
					ToPort:             int(toService.Spec.Ports[0].Port),
					ToNamespace:        toService.Namespace,
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
				table.Set(podKey(fromPod), serviceKey(toService), fmt.Sprintf("%d", curlExitCode))
			}
		}
	}

	table.Table().Render()
}
