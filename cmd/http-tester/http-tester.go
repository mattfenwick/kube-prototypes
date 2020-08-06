package main

import (
	httptester "github.com/mattfenwick/kube-prototypes/pkg/http-tester"
	"os"
)

func main() {
	httptester.Run(os.Args[1:])
}
