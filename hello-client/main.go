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
	hostAddr = "localhost:8080"
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
	// Use result to query Hello service
	target := fmt.Sprintf("http://%s/%s", addr, endpoint)
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