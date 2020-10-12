package main

import (
	"fmt"
	"github.com/mattfenwick/kube-prototypes/pkg/netpol"
)

func main() {
	ns := []string{"x", "y", "z"}
	pods := []string{"a", "b", "c"}
	var allPods []netpol.Pod
	for _, n := range ns {
		for _, p := range pods {
			allPods = append(allPods, netpol.NewPod(n, p))
		}
	}
	reachability := netpol.NewReachability(allPods, true)
	reachability.ExpectAllIngress("x/a", false)
	reachability.AllowLoopback()

	fmt.Println(reachability.Expected.PrettyPrint())

	// now update our matrix - we want anything 'y' to be able to get to x/a...
	//reachability.ExpectPeer(&netpol.Peer{Pod:"y"}, &netpol.Peer{}, true)
	//reachability.ExpectPeer(&netpol.Peer{Namespace:"y"}, &netpol.Peer{}, true)
	reachability.ExpectPeer(&netpol.Peer{}, &netpol.Peer{Namespace: "x"}, false)

	fmt.Println(reachability.Expected.PrettyPrint())
}
