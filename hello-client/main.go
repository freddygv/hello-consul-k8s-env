package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	endpoint = "hello"
	hostAddr = "hello.default.svc.cluster.local:8080"
	interval = 2 * time.Second
)

func main() {
	var (
		loop      = flag.Bool("loop", true, "Make continuous requests to hello service.")
	)
	flag.Parse()

	ticker := time.NewTicker(interval)
	for {
		if err := requestHello(hostAddr); err != nil {
			log.Printf("[ERR] failed to dial hello service: %v", err)
		}
		if !*loop {
			// Only run once if not looping
			break
		}
		<-ticker.C
	}
}

func requestHello(addr string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	target := fmt.Sprintf("http://%s/%s", addr, endpoint)
	resp, err := client.Get(target)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer client.CloseIdleConnections()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	log.Println(fmt.Sprintf("%s says: %s", target, body))
	return nil
}