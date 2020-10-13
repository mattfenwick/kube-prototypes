package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/kube"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/utils"
	"os"
	"strconv"
)

func main() {
	k, err := kube.NewKubernetes()
	utils.DoOrDie(err)

	ns, pod, cont, addr, portString := os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]
	port, err := strconv.Atoi(portString)
	utils.DoOrDie(err)

	for _, commandType := range []kube.ProbeCommandType{kube.ProbeCommandTypeCurl, kube.ProbeCommandTypeNetcat} {

		job := &kube.ProbeJob{
			FromNamespace:  ns,
			FromPod:        pod,
			FromContainer:  cont,
			ToAddress:      addr,
			ToPort:         port,
			TimeoutSeconds: 1,
			CommandType:    commandType,
			//FromKey:        "",
			//ToKey:          "",
		}
		result, err := k.Probe(job)
		utils.DoOrDie(err)

		bytes, err := json.MarshalIndent(result, "", "  ")
		utils.DoOrDie(err)
		fmt.Printf("result: %+v\n\n", string(bytes))
	}
}
