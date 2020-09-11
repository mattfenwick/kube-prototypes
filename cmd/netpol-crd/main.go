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
			Name:      "allow-nothing",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: selector,
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

func main() {
	clusterName := "calico-netpol"
	pathToKindSetupScript := "./kind-calico.sh"

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
	namespaceList := []string{"d1", "d2", "d3"}
	k8s, err := kube.NewKubernetes()
	utils.DoOrDie(err)
	for _, ns := range namespaceList {
		_, err = k8s.CreateOrUpdateNamespace(ns, map[string]string{"netpol-ns": ns})
		utils.DoOrDie(err)
		created, err := k8s.CreateDaemonSetIfNotExists(ns, kube.SimpleDaemonSet())
		utils.DoOrDie(err)
		_, err = k8s.CreateService(ns, kube.SimpleService())
		// give the daemonset some time to come up
		if created != nil {
			time.Sleep(10 * time.Second)
		}
	}

	// 4. run some probes
	initialResults, err := k8s.ProbePodToPod(namespaceList, 2)
	utils.DoOrDie(err)
	initialResults.Table().Render()

	// 5. install a few netpols
	utils.DoOrDie(k8s.CleanNetworkPolicies("d1"))
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
	_, err = k8s.CreateNetworkPolicy(AllowNothingFrom("d1", metav1.LabelSelector{}))
	utils.DoOrDie(err)

	// 6. probe again
	secondResults, err := k8s.ProbePodToPod(namespaceList, 2)
	utils.DoOrDie(err)
	secondResults.Table().Render()

	// 7. make a nice visualization of netpols
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
