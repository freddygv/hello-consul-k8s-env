package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	endpoint = "hello"
	hostname = "hello.service.consul"
	hostPort = "8080"
	interval = 2 * time.Second
)

func main() {
	var (
		loop      = flag.Bool("loop", true, "Make continuous requests to hello service.")
	)
	flag.Parse()

	ticker := time.NewTicker(interval)
	for {
		if err := requestHello(); err != nil {
			log.Printf("[ERR] failed to dial hello service: %v", err)
		}
		if !*loop {
			// Only run once if not looping
			break
		}
		<-ticker.C
	}
}

func requestHello() error {
	ips, err := net.LookupIP(hostname)
	if err != nil || len(ips) == 0 {
		return fmt.Errorf("could not find IP for '%s': %v", hostname, err)
	}

	// Use first result since they are shuffled by Consul
	addr := ips[0].String()

	// Use result to query Hello service
	target := fmt.Sprintf("http://%s/%s", net.JoinHostPort(addr, hostPort), endpoint)
	resp, err := http.Get(target)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	log.Println(fmt.Sprintf("%s says: %s", target, body))
	return nil
}