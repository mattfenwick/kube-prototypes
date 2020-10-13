package netpol_kube

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetpolServer struct {
	Name string
}

func (ns *NetpolServer) SimpleDaemonSet() *appsv1.DaemonSet {
	name := ns.Name
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

func (ns *NetpolServer) SimpleService() *v1.Service {
	name := ns.Name
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"component": name},
			Name:   name,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "port-7890", Port: 7890},
			},
			Selector: map[string]string{"component": name},
		},
		Status: v1.ServiceStatus{},
	}
}
