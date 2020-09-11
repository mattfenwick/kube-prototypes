package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/crd"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/examples"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/kube"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"time"
)


func AllowNothingFrom(namespace string, selector metav1.LabelSelector) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-nothing-from-%s", namespace),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: selector,
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

func AllowFromTo(namespace string, selector metav1.LabelSelector, nsLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-from-%s-to-%s", namespace, examples.LabelString(nsLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: selector,
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels:      nsLabels,
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	clusterName := "calico-netpol"
	pathToKindSetupScript := "./kind-calico.sh"
	k8s, err := kube.NewKubernetes()
	utils.DoOrDie(err)
	utils.DoOrDie(k8s.CleanNetworkPolicies("d1"))
	utils.DoOrDie(k8s.CleanNetworkPolicies("d2"))

	// 0. spin up a KinD cluster
	kindClient := kube.NewKindClient()
	exists, err := kindClient.GetCluster(clusterName)
	utils.DoOrDie(err)
	if !exists {
		log.Infof("cluster %s not found, creating ...", clusterName)
		cmd := exec.Command(pathToKindSetupScript, clusterName)
		utils.DoOrDie(utils.CommandRunAndPrint(cmd))
	} else {
		log.Infof("cluster %s already found, skipping creation ...", clusterName)
	}

	// 1. translate kube -> new
	log.Infof("converting policies from kube to new format:")
	kubeToNew()

	// 2. translate new -> kube
	log.Infof("converting policies from new format to kube:")
	newToKube()

	// 3. install some daemonsets
	namespaceList := []string{"d1", "d2"}//, "d3"}
	netpolDs := &kube.NetpolServer{Name: "netpol"}
	for _, ns := range namespaceList {
		_, err = k8s.CreateOrUpdateNamespace(ns, map[string]string{"netpol-ns": ns})
		utils.DoOrDie(err)
		created, err := k8s.CreateDaemonSetIfNotExists(ns, netpolDs.SimpleDaemonSet())
		utils.DoOrDie(err)
		_, err = k8s.CreateServiceIfNotExists(ns, netpolDs.SimpleService())
		utils.DoOrDie(err)
		// give the daemonset some time to come up
		if created != nil {
			time.Sleep(10 * time.Second)
		}
	}

	// 4. run some probes
	initialResults, err := k8s.ProbePodToPod(namespaceList, 2)
	utils.DoOrDie(err)
	fmt.Println("initial results:")
	initialResults.Table().Render()

	// 5. install a few netpols
	pols := []*crd.Policy{
		// TODO these policies don't work right, kube corner cases are hard to work with
		//   to deny, they should: select stuff in the target, and select *nothing* in
		//   the peers
		//crd.DenyEgressFromNamespace("d1"),
		//crd.DenyAll,
	}
	for _, pol := range pols {
		netpols := crd.Reduce(pol)
		for _, netpol := range netpols {
			nYaml, err := yaml.Marshal(netpol)
			utils.DoOrDie(err)
			log.Infof("creating policy: \n\n%s\n\n", nYaml)
			_, err = k8s.CreateNetworkPolicy(netpol)
			utils.DoOrDie(err)
		}
	}

	_, err = k8s.CreateNetworkPolicy(crd.AllowAllIngressNetworkingPolicy("d2"))
	utils.DoOrDie(err)
	allowAllToD2, err := k8s.ProbePodToPod(namespaceList, 2)
	utils.DoOrDie(err)
	fmt.Println("allow-all-to-d2 results:")
	allowAllToD2.Table().Render()

	_, err = k8s.CreateNetworkPolicy(AllowNothingFrom("d1", metav1.LabelSelector{}))
	utils.DoOrDie(err)

	// 6. probe again
	secondResults, err := k8s.ProbePodToPod(namespaceList, 2)
	utils.DoOrDie(err)
	fmt.Println("deny-all-from-d1 results:")
	secondResults.Table().Render()

	// 7. allow some ingress
	_, err = k8s.CreateNetworkPolicy(AllowFromTo("d1", metav1.LabelSelector{}, map[string]string{"netpol-ns": "d2"}))
	utils.DoOrDie(err)
	//_, err = k8s.CreateNetworkPolicy(crd.AllowAllIngressNetworkingPolicy("d2"))
	//utils.DoOrDie(err)

	// 8. probe again
	thirdResults, err := k8s.ProbePodToPod(namespaceList, 2)
	utils.DoOrDie(err)
	fmt.Println("allow-all-to-d2 results:")
	thirdResults.Table().Render()

	// 9. make a nice visualization of netpols
	panic("TODO")
}

func kubeToNew() {
	// kube -> this
	netpol := examples.AllowFromNamespaceTo(
		"abcd",
		map[string]string{"purpose": "production"},
		map[string]string{"app": "web"})
	policies := crd.BuildTarget(netpol)
	bytes, err := json.MarshalIndent(policies, "", "  ")
	utils.DoOrDie(err)
	fmt.Printf("%s\n\n", bytes)

	yamlBytes, err := yaml.Marshal(policies)
	utils.DoOrDie(err)
	fmt.Printf("%s\n\n", yamlBytes)
}

func newToKube() {
	// this -> kube
	np := crd.DenyAll
	kubeNetPols := crd.Reduce(np)
	bytes, err := json.MarshalIndent(kubeNetPols, "", "  ")
	utils.DoOrDie(err)
	fmt.Printf("%s\n\n", bytes)

	yamlBytes, err := yaml.Marshal(kubeNetPols)
	utils.DoOrDie(err)
	fmt.Printf("%s\n\n", yamlBytes)
}
