package kube

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) CreateDaemonSet(namespace string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	return k.ClientSet.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
}

func (k *Kubernetes) CreateDaemonSetIfNotExists(namespace string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	created, err := k.ClientSet.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
	if err == nil {
		return created, nil
	}
	if err.Error() == fmt.Sprintf(`daemonsets.apps "%s" already exists`, ds.Name) {
		return nil, nil
	}
	return nil, err
}

func (k *Kubernetes) CreateService(namespace string, svc *v1.Service) (*v1.Service, error) {
	return k.ClientSet.CoreV1().Services(namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
}

func SimpleDaemonSet() *appsv1.DaemonSet {
	name := "netpol"
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"component": name,
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": name},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"name": name, "component": name},
				},
				Spec: v1.PodSpec{
					Tolerations: []v1.Toleration{
						// this toleration is to have the daemonset runnable on master nodes
						// remove it if your masters can't run pods
						{Key: "node-role.kubernetes.io/master", Effect: v1.TaintEffectNoSchedule},
					},
					Containers: []v1.Container{
						{
							Name:    name,
							Image:   "docker.io/mfenwick100/kube-prototypes-client-server:latest",
							Command: []string{"./http-tester"},
							Args:    []string{"server", "7890"},
							Ports: []v1.ContainerPort{
								{ContainerPort: 7890, Protocol: v1.ProtocolTCP},
							},
							//Resources: struct {
							//	Limits   v1.ResourceList
							//	Requests v1.ResourceList
							//}{Limits:, Requests:},
							ImagePullPolicy: v1.PullAlways,
						},
					},
				},
			},
		},
	}
}

func SimpleService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"component": "netpol-server"},
			Name:   "netpol-server",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "port-7890", Port: 7890},
			},
			Selector: map[string]string{"component": "netpol-server"},
		},
		Status: v1.ServiceStatus{},
	}
}
