package kube

import (
	"context"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PodKey(pod v1.Pod) string {
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}

func GetPods(k8s *Kubernetes, namespaces []string) ([]v1.Pod, error) {
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

func (k *Kubernetes) ProbePodToPod(namespaces []string, timeoutSeconds int) (*netpol.StringTruthTable, error) {
	pods, err := GetPods(k, namespaces)
	if err != nil {
		return nil, err
	}

	var items []string
	for _, p := range pods {
		items = append(items, PodKey(p))
	}

	table := netpol.NewStringTruthTable(items)
	for _, fromPod := range pods {
		if len(fromPod.Spec.Containers) != 1 || len(fromPod.Status.ContainerStatuses) != 1 {
			panic(errors.Errorf("fromPod: expected 1 container spec and 1 container status, found %+v", fromPod))
		}
		fromContainer := fromPod.Spec.Containers[0].Name
		for _, toPod := range pods {
			if len(toPod.Spec.Containers) != 1 || len(toPod.Status.ContainerStatuses) != 1 {
				panic(errors.Errorf("toPod: expected 1 container spec and 1 container status, found %+v", toPod))
			}
			toCont := toPod.Spec.Containers[0]
			log.Infof("Probing %s -> %s", PodKey(fromPod), PodKey(toPod))
			//connected, err := k.ProbeWithPod(fromPod, toPod, port)
			if len(toCont.Ports) == 0 {
				table.Set(PodKey(fromPod), PodKey(toPod), "no ports")
				continue
			}
			connected, curlExitCode, err := k.ProbeFromContainerToPod(&ProbeFromContainerToPod{
				FromNamespace:      fromPod.Namespace,
				FromPod:            fromPod.Name,
				FromContainer:      fromContainer,
				ToIP:               toPod.Status.PodIP,
				ToPort:             int(toCont.Ports[0].ContainerPort),
				ToNamespace:        toPod.Namespace,
				ToPod:              toPod.Name,
				CurlTimeoutSeconds: timeoutSeconds,
			})
			log.Warningf("curl exit code: %d", curlExitCode)
			if err != nil {
				log.Errorf("unable to make main observation on %s -> %s: %+v", fromPod.Name, toPod.Name, err)
			}
			if !connected {
				log.Warnf("FAILED CONNECTION FOR WHITELISTED PODS %s -> %s !!!! ", fromPod.Name, toPod.Name)
			}
			table.Set(PodKey(fromPod), PodKey(toPod), fmt.Sprintf("%d", curlExitCode))
		}
	}

	return table, nil
}
