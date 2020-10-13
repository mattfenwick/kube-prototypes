package kube

import (
	"bytes"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"regexp"
	"strconv"
	"strings"
)

type ProbeJob struct {
	FromNamespace      string
	FromPod            string
	FromContainer      string
	ToAddress          string
	ToPort             int
	CurlTimeoutSeconds int
	FromKey            string
	ToKey              string
}

func (pj *ProbeJob) GetFromKey() string {
	if pj.FromKey != "" {
		return pj.FromKey
	}
	return fmt.Sprintf("%s/%s/%s", pj.FromNamespace, pj.FromPod, pj.FromContainer)
}

func (pj *ProbeJob) GetToKey() string {
	if pj.ToKey != "" {
		return pj.ToKey
	}
	return fmt.Sprintf("%s:%d", pj.ToAddress, pj.ToPort)
}

func (pj *ProbeJob) KubeExecCommand() []string {
	return append([]string{
		"kubectl", "exec",
		pj.FromPod,
		"-c", pj.FromContainer,
		"-n", pj.FromNamespace,
		"--",
	},
		pj.CurlCommand()...)
}

func (pj *ProbeJob) ToURL() string {
	return fmt.Sprintf("http://%s:%d", pj.ToAddress, pj.ToPort)
}

//func (pj *ProbeJob)WGetCommand() []string {
//	// note some versions of wget want -s for spider mode, others, -S
//	return []string{"wget", "--spider", "--tries", "1", "--timeout", "1", pj.ToURL()}
//}

func (pj *ProbeJob) CurlCommand() []string {
	return []string{"curl", "-I", "--connect-timeout", fmt.Sprintf("%d", pj.CurlTimeoutSeconds), pj.ToURL()}
}

func (k *Kubernetes) Probe(job *ProbeJob) (bool, int, error) {
	fromNamespace := job.FromNamespace
	fromPod := job.FromPod
	fromContainer := job.FromContainer

	curlRegexp := regexp.MustCompile(`curl: \((\d+)\)`)

	log.Infof("Running: %s", strings.Join(job.KubeExecCommand(), " "))
	out, errorOut, err := k.ExecuteRemoteCommand(fromNamespace, fromPod, fromContainer, job.CurlCommand())
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

	log.Errorf("failed connect:\n - %v\n - %v\n - %s/%s", out, errorOut, fromNamespace, fromPod)
	return false, -1, err
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

type ProbeJobResults struct {
	Job         *ProbeJob
	IsConnected bool
	ExitCode    int
	Err         error
}

func (k *Kubernetes) ProbeConnectivity(jobs []*ProbeJob) *netpol.StringTruthTable {
	numberOfWorkers := 30
	jobsChan := make(chan *ProbeJob, len(jobs))
	results := make(chan *ProbeJobResults, len(jobs))
	for i := 0; i < numberOfWorkers; i++ {
		go probeWorker(k, jobsChan, results)
	}
	var froms, tos []string
	fromSet := map[string]bool{}
	toSet := map[string]bool{}
	for _, job := range jobs {
		jobsChan <- job
		if _, ok := fromSet[job.GetFromKey()]; !ok {
			froms = append(froms, job.GetFromKey())
			fromSet[job.GetFromKey()] = true
		}
		if _, ok := toSet[job.GetToKey()]; !ok {
			tos = append(tos, job.GetToKey())
			toSet[job.GetToKey()] = true
		}
	}
	close(jobsChan)

	table := netpol.NewStringTruthTableWithFromsTo(froms, tos)

	for i := 0; i < len(jobs); i++ {
		result := <-results
		job := result.Job
		if result.Err != nil {
			log.Infof("unable to perform probe %s/%s/%s -> %s:%d : %v", job.FromNamespace, job.FromPod, job.FromContainer, job.ToAddress, job.ToPort, result.Err)
		}
		table.Set(job.GetFromKey(), job.GetToKey(), fmt.Sprintf("%t", result.IsConnected))
	}
	return table
}

func probeWorker(k8s *Kubernetes, jobs <-chan *ProbeJob, results chan<- *ProbeJobResults) {
	for job := range jobs {
		log.Debugf("starting probe job %+v", job)
		connected, exitCode, err := k8s.Probe(job)
		result := &ProbeJobResults{
			Job:         job,
			IsConnected: connected,
			ExitCode:    exitCode,
			Err:         err,
		}
		log.Debugf("finished probe job %+v", result)
		results <- result
	}
}

// convenience functions

func (k *Kubernetes) ProbePodToPod(namespaces []string, timeoutSeconds int) (*netpol.StringTruthTable, error) {
	pods, err := k.GetPodsInNamespaces(namespaces)
	if err != nil {
		return nil, err
	}

	var jobs []*ProbeJob
	// TODO could verify that each pod only has one container, each container only has one port, etc.
	for _, from := range pods {
		for _, to := range pods {
			jobs = append(jobs, &ProbeJob{
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

	return k.ProbeConnectivity(jobs), nil
}
