package kube

import (
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/utils"
	"github.com/pkg/errors"
	"io/ioutil"
	"os/exec"
	"strings"
)

type KindClient struct{}

func NewKindClient() *KindClient {
	return &KindClient{}
}

func (kc *KindClient) GetVersion() (string, error) {
	return utils.CommandRun(exec.Command("kind", "version"))
}

func (kc *KindClient) GetClusters() ([]string, error) {
	output, err := utils.CommandRun(exec.Command("kind", "get", "clusters"))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(output, "\n")
	// drop the last line, if it's empty
	var clusters []string
	for _, line := range lines {
		if len(line) > 0 {
			clusters = append(clusters, line)
		}
	}
	// wish there were a --json option on KinD
	if len(clusters) == 1 && clusters[0] == "No kind clusters found." {
		return []string{}, nil
	}
	return clusters, nil
}

func (kc *KindClient) GetCluster(name string) (bool, error) {
	clusters, err := kc.GetClusters()
	if err != nil {
		return false, err
	}
	for _, cluster := range clusters {
		if name == cluster {
			return true, nil
		}
	}
	return false, nil
}

func (kc *KindClient) CreateCluster(clusterName string, image string, config string) error {
	configFilePath := fmt.Sprintf("polaris-local-kind-config-%s.yaml", clusterName)
	err := ioutil.WriteFile(configFilePath, []byte(config), 0755)
	if err != nil {
		return errors.Wrapf(err, "unable to write kind config file %s", configFilePath)
	}
	return kc.CreateClusterWithConfigFile(clusterName, image, configFilePath)
}

func (kc *KindClient) CreateClusterWithConfigFile(clusterName string, image string, configFilePath string) error {
	// kind create cluster --name "${KIND_CLUSTER_NAME}" --image "${KIND_NODE_IMAGE}" --config $CONFIG_FILE_PATH
	cmd := exec.Command("kind",
		"create", "cluster",
		"--name", clusterName,
		"--image", image,
		"--config", configFilePath)
	return utils.CommandRunAndPrint(cmd)
}
