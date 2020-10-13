package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/kube"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/crd"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/matcher"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/netpol-kube/examples"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpLogger(logLevelStr string) error {
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		return errors.Wrapf(err, "unable to parse the specified log level: '%s'", logLevel)
	}
	log.SetLevel(logLevel)
	//log.SetFormatter(&log.TextFormatter{
	//	FullTimestamp: true,
	//})
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
	command.AddCommand(SetupCreateNetpolCommand())

	return command
}

func SetupCreateNetpolCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "create some netpols",
		Long:  "create some netpols",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// this -> kube
			np := crd.DenyAll
			kubeNetPols := crd.Reduce(np)
			bytes, err := json.MarshalIndent(kubeNetPols, "", "  ")
			utils.DoOrDie(err)
			fmt.Printf("%s\n\n", bytes)

			yamlBytes, err := yaml.Marshal(kubeNetPols)
			utils.DoOrDie(err)
			fmt.Printf("%s\n\n", yamlBytes)

			// kube -> this
			netpol := examples.AllowFromNamespaceTo(
				"abcd",
				map[string]string{"purpose": "production"},
				map[string]string{"app": "web"})
			policies := crd.BuildTarget(netpol)
			bytes, err = json.MarshalIndent(policies, "", "  ")
			utils.DoOrDie(err)
			fmt.Printf("%s\n\n", bytes)

			yamlBytes, err = yaml.Marshal(policies)
			utils.DoOrDie(err)
			fmt.Printf("%s\n\n", yamlBytes)
		},
	}

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

type ProbeArgs struct {
	Namespaces     []string
	ProbePods      bool
	ProbeServices  bool
	TimeoutSeconds int
}

func SetupProbeCommand() *cobra.Command {
	args := &ProbeArgs{}

	command := &cobra.Command{
		Use:   "probe",
		Short: "probe connectivity",
		Long:  "probe pod -> pod and pod -> service connectivity",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			runProbe(args)
		},
	}

	command.Flags().StringSliceVar(&args.Namespaces, "nss", []string{}, "namespaces to probe")
	command.MarkFlagRequired("nss")

	command.Flags().BoolVar(&args.ProbePods, "pod", true, "probe pods")
	command.Flags().BoolVar(&args.ProbeServices, "svc", false, "probe services")

	command.Flags().IntVar(&args.TimeoutSeconds, "timeout", 2, "timeout in seconds")

	return command
}

func main() {
	command := setupCommand()
	err := errors.Wrapf(command.Execute(), "run root command")
	utils.DoOrDie(err)

	// 1. probe all pods in a namespace
	//  - on a few different ports ?  or on the port opened for the services?  <- get the right ports
	// 2. apply a blanket-deny netpol
	//  - probe pods again
	// 3. start opening communication between pods
	// 4. figure out some visualization of connectivity
}

func mungeNetworkPolicies() {
	k8s, err := kube.NewKubernetes()
	utils.DoOrDie(err)

	err = k8s.CleanNetworkPolicies("default")
	utils.DoOrDie(err)

	var allCreated []*networkingv1.NetworkPolicy
	for _, np := range examples.AllExamples {
		createdNp, err := k8s.CreateNetworkPolicy(np)
		allCreated = append(allCreated, createdNp)
		utils.DoOrDie(err)
		//explanation := netpol.ExplainPolicy(np)
		explanation := netpol.ExplainPolicy(createdNp)
		fmt.Printf("policy explanation for %s:\n%s\n\n", np.Name, explanation.PrettyPrint())

		matcherExplanation := matcher.Explain(matcher.BuildNetworkPolicy(createdNp))
		fmt.Printf("\nmatcher explanation: %s\n\n", matcherExplanation)

		reduced := netpol.Reduce(createdNp)
		fmt.Println(netpol.NodePrettyPrint(reduced))
		fmt.Println()

		createdNpBytes, err := json.MarshalIndent(createdNp, "", "  ")
		utils.DoOrDie(err)
		fmt.Printf("created netpol:\n\n%s\n\n", createdNpBytes)

		matcherPolicy := matcher.BuildNetworkPolicy(createdNp)
		matcherPolicyBytes, err := json.MarshalIndent(matcherPolicy, "", "  ")
		utils.DoOrDie(err)
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
	utils.DoOrDie(err)
	fmt.Printf("full network policies:\n\n%s\n\n", bytes)
	fmt.Printf("\nexplained:\n%s\n", matcher.Explain(netpols))

	netpolsExamples := matcher.BuildNetworkPolicy(examples.ExampleComplicatedNetworkPolicy())
	fmt.Printf("complicated example explained:\n%s\n", matcher.Explain(netpolsExamples))
}

func runProbe(args *ProbeArgs) {
	k8s, err := kube.NewKubernetes()
	utils.DoOrDie(err)

	//if args.ProbeContainers {
	//	probeContainerToContainer(args.Namespaces, k8s, args.TimeoutSeconds)
	//}

	if args.ProbePods {
		probePodToPod(args.Namespaces, k8s, args.TimeoutSeconds)
	}

	if args.ProbeServices {
		probeContainerToService(args.Namespaces, k8s, args.TimeoutSeconds)
	}
}

func serviceKey(svc v1.Service) string {
	return fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
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

func probePodToPod(namespaces []string, k8s *kube.Kubernetes, timeoutSeconds int) {
	pods, err := k8s.GetPodsInNamespaces(namespaces)
	utils.DoOrDie(err)

	var jobs []*kube.ProbeJob
	// TODO could verify that each pod only has one container, each container only has one port, etc.
	for _, from := range pods {
		for _, to := range pods {
			jobs = append(jobs, &kube.ProbeJob{
				FromNamespace:      from.Namespace,
				FromPod:            from.Name,
				FromContainer:      from.Spec.Containers[0].Name,
				ToAddress:          to.Status.PodIP,
				ToPort:             int(to.Spec.Containers[0].Ports[0].ContainerPort),
				CurlTimeoutSeconds: timeoutSeconds,
				// TODO
				//FromKey:            "",
				//ToKey:              "",
			})
		}
	}

	table := k8s.ProbeConnectivity(jobs)

	table.Table().Render()
}

func probeContainerToService(namespaces []string, k8s *kube.Kubernetes, timeoutSeconds int) {
	pods, err := k8s.GetPodsInNamespaces(namespaces)
	utils.DoOrDie(err)

	services, err := getServices(k8s, namespaces)
	utils.DoOrDie(err)

	var jobs []*kube.ProbeJob
	for _, p := range pods {
		for _, c := range p.Spec.Containers {
			for _, s := range services {
				jobs = append(jobs, &kube.ProbeJob{
					FromNamespace:      p.Namespace,
					FromPod:            p.Name,
					FromContainer:      c.Name,
					ToAddress:          kube.QualifiedServiceAddress(s.Namespace, s.Name),
					ToPort:             int(s.Spec.Ports[0].Port),
					CurlTimeoutSeconds: timeoutSeconds,
					// TODO
					//FromKey:            "",
					//ToKey:              "",
				})
			}
		}
	}

	table := k8s.ProbeConnectivity(jobs)

	table.Table().Render()
}
