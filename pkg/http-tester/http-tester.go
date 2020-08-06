package http_tester

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Run(args []string) {
	if args[0] == "client" {
		RunClient(args[1:])
	} else if args[0] == "server" {
		RunServer(args[1:])
	} else {
		panic(errors.Errorf("expected server or client for first arg, found %s", args[0]))
	}
}
