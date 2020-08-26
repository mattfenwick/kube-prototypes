package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

type Results struct {
	IP4           net.IP
	IP16          net.IP
	IsLoopback    bool
	Mask          net.IPMask
	MaskString    string
	IPNetIPString string
	IPNetString   string
	IPNetNetwork  string
}

func IPNetToResults(ipnet *net.IPNet) *Results {
	net.ParseCIDR()
	return &Results{
		IP4:           ipnet.IP.To4(),
		IP16:          ipnet.IP.To16(),
		IsLoopback:    ipnet.IP.IsLoopback(),
		Mask:          ipnet.Mask,
		MaskString:    ipnet.Mask.String(),
		IPNetIPString: ipnet.IP.String(),
		IPNetString:   ipnet.String(),
		IPNetNetwork:  ipnet.Network(),
	}
}

func main() {
	addrs, err := net.InterfaceAddrs()
	doOrDie(err)

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			//fmt.Printf("*net.IPNet: %s, %s, %t, %s, %s\n%s, %s\n\n", ipnet.IP.To4(), ipnet.IP.To16(), ipnet.IP.IsLoopback(), ipnet.Mask, ipnet.IP.String(), ipnet.String(), ipnet.Network())
			results := IPNetToResults(ipnet)
			bytes, err := json.MarshalIndent(results, "", "  ")
			doOrDie(err)
			fmt.Printf("%s\n\n", bytes)
			//fmt.Printf("%+v\n\n", results)
		} else {
			fmt.Printf("something else: %+v, %s, %s, %T\n", a, a.String(), a.Network(), a)
		}
	}

	interfaces, err := net.Interfaces()
	doOrDie(err)
	fmt.Printf("interfaces:\n\n")
	for _, i := range interfaces {
		fmt.Printf("%+v, %T\n\n", i, i)
	}

	// TODO this doesn't work offline
	if false {
		primary()
	}
}

func primary() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	doOrDie(err)

	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Printf("local addr: %+v\n", localAddr)
}
