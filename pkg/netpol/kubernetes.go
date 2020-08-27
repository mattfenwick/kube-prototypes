package netpol

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type Kubernetes struct {
	podCache  map[string][]v1.Pod
	ClientSet *kubernetes.Clientset
}

func NewKubernetes() (*Kubernetes, error) {
	clientSet, err := Clientset()
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to instantiate kube client")
	}
	return &Kubernetes{
		podCache:  map[string][]v1.Pod{},
		ClientSet: clientSet,
	}, nil
}

func Clientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to build config from flags, check that your KUBECONFIG file is correct !")
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate clientset")
	}
	return clientset, nil
}

//func (k *Kubernetes) GetPods(ns string, key string, val string) ([]v1.Pod, error) {
//	if p, ok := k.podCache[fmt.Sprintf("%v_%v_%v", ns, key, val)]; ok {
//		return p, nil
//	}
//
//	v1PodList, err := k.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{})
//	if err != nil {
//		return nil, errors.Wrapf(err, "unable to list pods")
//	}
//	pods := []v1.Pod{}
//	for _, pod := range v1PodList.Items {
//		// log.Infof("check: %s, %s, %s, %s", pod.Name, pod.Labels, key, val)
//		if pod.Labels[key] == val {
//			pods = append(pods, pod)
//		}
//	}
//
//	//log.Infof("list in ns %s: %d -> %d", ns, len(v1PodList.Items), len(pods))
//	k.podCache[fmt.Sprintf("%v_%v_%v", ns, key, val)] = pods
//
//	return pods, nil
//}

//func (k *Kubernetes) GetPod(ns string, key string, val string) (v1.Pod, error) {
//	pods := k.GetPods(ns, key, val)
//	if len(pods) != 1 {
//		return errors.Errorf("expected 1 pod of ")
//	}
//}

//func (k *Kubernetes) Probe(ns1 string, pod1 string, ns2 string, pod2 string, port int) (bool, error) {
//	fromPods, err := k.GetPods(ns1, "pod", pod1)
//	if err != nil {
//		return false, errors.WithMessagef(err, "unable to get pods from ns %s", ns1)
//	}
//	if len(fromPods) == 0 {
//		return false, errors.Errorf("no pod of name %s in namespace %s found", pod1, ns1)
//	}
//	fromPod := fromPods[0]
//
//	toPods, err := k.GetPods(ns2, "pod", pod2)
//	if err != nil {
//		return false, errors.WithMessagef(err, "unable to get pods from ns %s", ns2)
//	}
//	if len(toPods) == 0 {
//		return false, errors.New(fmt.Sprintf("no pod of name %s in namespace %s found", pod2, ns2))
//	}
//	toPod := toPods[0]
//
//	toIP := toPod.Status.PodIP
//
//	// note some versions of wget want -s for spider mode, others, -S
//	exec := []string{"wget", "--spider", "--tries", "1", "--timeout", "1", "http://" + toIP + ":" + fmt.Sprintf("%v", port)}
//
//	containerName := fromPod.Status.ContainerStatuses[0].Name
//	log.Info("Running: kubectl exec -t -i " + fromPod.Name + " -c " + containerName + " -n " + fromPod.Namespace + " -- " + strings.Join(exec, " "))
//	out, out2, err := k.ExecuteRemoteCommand(fromPod, containerName, exec)
//	log.Info(".... Done")
//
//	if err != nil {
//		log.Errorf("failed connect.... %v %v %v %v %v %v", out, out2, ns1, pod1, ns2, pod2)
//		return false, errors.WithMessagef(err, "unable to execute remote command %+v", exec)
//	}
//	return true, nil
//}

func (k *Kubernetes) ProbeWithPod(podFrom v1.Pod, podTo v1.Pod, port int) (bool, error) {
	toIP := podTo.Status.PodIP

	// note some versions of wget want -s for spider mode, others, -S
	address := fmt.Sprintf("http://%s:%d", toIP, port)
	//exec := []string{"wget", "--spider", "--tries", "1", "--timeout", "1", address}
	exec := []string{"curl", "-I", address}

	containerName := podFrom.Status.ContainerStatuses[0].Name
	log.Info("Running: kubectl exec -t -i " + podFrom.Name + " -c " + containerName + " -n " + podFrom.Namespace + " -- " + strings.Join(exec, " "))
	out, errorOut, err := k.ExecuteRemoteCommand(podFrom.Namespace, podFrom.Name, containerName, exec)
	log.Info(".... Done")

	if err != nil {
		log.Errorf("failed connect:\n - %v\n - %v\n - %v\n - %v\n - %v\n - %v", out, errorOut, podFrom.Namespace, podFrom.Name, podTo.Namespace, podTo.Name)
		return false, errors.WithMessagef(err, "unable to execute remote command %+v", exec)
	}
	return true, nil
}

type ProbeFromContainerToPod struct {
	FromNamespace      string
	FromPod            string
	FromContainer      string
	ToIP               string
	ToPort             int
	ToNamespace        string
	ToPod              string
	CurlTimeoutSeconds int
}

func (k *Kubernetes) ProbeFromContainerToPod(args *ProbeFromContainerToPod) (bool, int, error) {
	fromNamespace := args.FromNamespace
	fromPod := args.FromPod
	fromContainer := args.FromContainer
	toIP := args.ToIP
	toPort := args.ToPort
	toPod := args.ToPod
	toNamespace := args.ToNamespace
	curlTimeoutSeconds := args.CurlTimeoutSeconds

	address := fmt.Sprintf("http://%s:%d", toIP, toPort)
	// note some versions of wget want -s for spider mode, others, -S
	//exec := []string{"wget", "--spider", "--tries", "1", "--timeout", "1", address}
	exec := []string{"curl", "-I", "--connect-timeout", fmt.Sprintf("%d", curlTimeoutSeconds), address}

	curlRegexp := regexp.MustCompile(`curl: \((\d+)\)`)

	log.Infof("Running: kubectl exec -t -i %s -c %s -n %s -- %s", fromPod, fromContainer, fromNamespace, strings.Join(exec, " "))
	out, errorOut, err := k.ExecuteRemoteCommand(fromNamespace, fromPod, fromContainer, exec)
	log.Infof("finished, with out '%s' and errOut '%s'", out, errorOut)

	if err == nil {
		matches := curlRegexp.FindStringSubmatch(out)
		if len(matches) == 0 {
			return true, 0, nil
		}
		curlExitCode, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatalf("unable to parse '%s' to int: %+v", matches[1], err)
		}
		return true, curlExitCode, nil
	}

	log.Errorf("failed connect:\n - %v\n - %v\n - %v\n - %v\n - %v\n - %v", out, errorOut, fromNamespace, fromPod, toNamespace, toPod)
	return false, -1, errors.WithMessagef(err, "unable to execute remote command %+v", exec)
}

// ExecuteRemoteCommand executes a remote shell command on the given pod
// returns the output from stdout and stderr
func (k *Kubernetes) ExecuteRemoteCommand(namespace string, pod string, container string, command []string) (string, string, error) {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to get rest config from kube config")
	}

	request := k.ClientSet.
		CoreV1().
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(
			&v1.PodExecOptions{
				Container: container,
				Command:   command,
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			},
			scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", request.URL())
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to instantiate SPDYExecutor")
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})

	out, errOut := buf.String(), errBuf.String()
	if err == nil {
		return out, errOut, nil
	}
	// Goal: distinguish between "we weren't able to run the command, for whatever reason" and
	//   "we were able to run the command just fine, but the command itself failed"
	// TODO Not sure if this is accomplishing that goal correctly
	if out != "" || errOut != "" {
		log.Warningf("ignoring error for command '%s' on %s/%s/%s: %+v", command, namespace, pod, container, err)
		return out, errOut, nil
	}
	return out, errOut, errors.Wrapf(err, "Unable to run command %s on %v/%v/%v", command, namespace, pod, container)
}

func (k *Kubernetes) CreateOrUpdateNamespace(n string, labels map[string]string) (*v1.Namespace, error) {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   n,
			Labels: labels,
		},
	}
	nsr, err := k.ClientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err == nil {
		log.Infof("created namespace %s", ns)
		return nsr, errors.Wrapf(err, "unable to create namespace %s", ns)
	}

	log.Debugf("unable to create namespace %s, let's try updating it instead (error: %s)", ns.Name, err)
	nsr, err = k.ClientSet.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		log.Debugf("unable to create namespace %s: %s", ns, err)
	}

	return nsr, err
}

func (k *Kubernetes) CreateOrUpdateDeployment(ns, deploymentName string, replicas int32, labels map[string]string) (*appsv1.Deployment, error) {
	zero := int64(0)
	log.Infof("creating/updating deployment %s in ns %s", deploymentName, ns)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Labels:    labels,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    labels,
					Namespace: ns,
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &zero,
					Containers: []v1.Container{
						{
							Name:            "c80",
							Image:           "python:latest",
							Command:         []string{"python", "-m", "http.server", "80"},
							SecurityContext: &v1.SecurityContext{},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "serve-80",
								},
							},
						},
						{
							Name:            "c81",
							Image:           "python:latest",
							Command:         []string{"python", "-m", "http.server", "81"},
							SecurityContext: &v1.SecurityContext{},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 81,
									Name:          "serve-81",
								},
							},
						},
					},
				},
			},
		},
	}

	d, err := k.ClientSet.AppsV1().Deployments(ns).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err == nil {
		log.Infof("created deployment %s in namespace %s", d.Name, ns)
		return d, nil
	}

	log.Debugf("unable to create deployment %s in ns %s, let's try update instead", deployment.Name, ns)
	d, err = k.ClientSet.AppsV1().Deployments(ns).Update(context.TODO(), d, metav1.UpdateOptions{})
	if err != nil {
		log.Debugf("unable to update deployment %s in ns %s: %s", deployment.Name, ns, err)
	}
	return d, err
}

func (k *Kubernetes) CleanNetworkPolicies(ns string) error {
	netpols, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to list network policies in ns %s", ns)
	}
	for _, np := range netpols.Items {
		log.Infof("deleting network policy %s in ns %s", np.Name, ns)
		err = k.ClientSet.NetworkingV1().NetworkPolicies(np.Namespace).Delete(context.TODO(), np.Name, metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrapf(err, "unable to delete netpol %s/%s", ns, np.Name)
		}
	}
	return nil
}

func (k *Kubernetes) CreateNetworkPolicy(netpol *v1net.NetworkPolicy) (*v1net.NetworkPolicy, error) {
	ns := netpol.Namespace
	log.Infof("creating network policy %s in ns %s", netpol.Name, ns)

	createdPolicy, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).Create(context.TODO(), netpol, metav1.CreateOptions{})
	return createdPolicy, errors.Wrapf(err, "unable to create network policy %s/%s", netpol.Name, netpol.Namespace)
}

func (k *Kubernetes) CreateOrUpdateNetworkPolicy(ns string, netpol *v1net.NetworkPolicy) (*v1net.NetworkPolicy, error) {
	log.Infof("creating/updating network policy %s in ns %s", netpol.Name, ns)
	netpol.ObjectMeta.Namespace = ns
	np, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).Update(context.TODO(), netpol, metav1.UpdateOptions{})
	if err == nil {
		return np, err
	}

	log.Debugf("unable to update network policy %s in ns %s, let's try creating it instead (error: %s)", netpol.Name, ns, err)
	np, err = k.ClientSet.NetworkingV1().NetworkPolicies(ns).Create(context.TODO(), netpol, metav1.CreateOptions{})
	if err != nil {
		log.Debugf("unable to create network policy: %s", err)
	}
	return np, err
}
