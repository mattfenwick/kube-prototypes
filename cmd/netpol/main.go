package main

import (
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
	log "github.com/sirupsen/logrus"
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

	ns := os.Args[1]
	k8s, err := netpol.NewKubernetes()
	doOrDie(err)
	validate(ns, k8s, 8443)
}

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func validate(ns string, k8s *netpol.Kubernetes, port int) {
	pods, err := k8s.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{})
	doOrDie(err)
	var items []string
	for _, p := range pods.Items {
		// the order of status containers doesn't necessarily match the order of spec containers for the same
		//   pod, so let's just always use the spec to be consistent and reduce confusion!
		for _, c := range p.Spec.Containers {
			items = append(items, c.Name)
		}
	}
	falseConstant := false
	table := netpol.NewTruthTable(items, &falseConstant)
	for _, fromPod := range pods.Items {
		for _, fromCont := range fromPod.Spec.Containers {
			fromContainer := fromCont.Name
			for _, toPod := range pods.Items {
				for _, toCont := range toPod.Spec.Containers {
					log.Infof("Probing in ns %s: %s, %s", ns, fromPod.Name, toPod.Name)
					//connected, err := k8s.ProbeWithPod(fromPod, toPod, port)
					connected, curlExitCode, err := k8s.ProbeFromContainerToPod(&netpol.ProbeFromContainerToPod{
						FromNamespace: fromPod.Namespace,
						FromPod:       fromPod.Name,
						FromContainer: fromContainer,
						ToIP:          toPod.Status.PodIP,
						ToPort:        int(toCont.Ports[0].ContainerPort),
						ToNamespace:   toPod.Namespace,
						ToPod:         toPod.Name,
					})
					log.Warningf("curl exit code: %d", curlExitCode)
					if err != nil {
						log.Errorf("unable to make main observation on %s -> %s: %+v", fromPod.Name, toPod.Name, err)
					}
					if !connected {
						log.Warnf("FAILED CONNECTION FOR WHITELISTED PODS %s -> %s !!!! ", fromPod.Name, toPod.Name)
					}
					table.Set(fromContainer, toCont.Name, connected)
				}
			}
		}
	}

	table.Table().Render()
}
