package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type serverConfig struct {
	mu           sync.RWMutex
	Language     *string   `json:"language"`
	ConsulAddr   *string   `json:"consul_addr"`
	KVPath       *string   `json:"kv_path"`
	ServiceName  *string   `json:"service_name"`
	TTLEndpoint  *string   `json:"ttl_endpoint"`
	TTLID        *string   `json:"ttl_id"`
	EnableChecks *bool     `json:"enable_checks"`
	DebugMode    *bool     `json:"debug_mode"`
	ToWatch      *[]string `json:"keys_to_watch"`
}

func (c *serverConfig) merge(other *serverConfig) *serverConfig {
	o := *other
	if c == nil {
		return &o
	}
	if c.Language == nil {
		c.Language = o.Language
	}
	if c.ConsulAddr == nil {
		c.ConsulAddr = o.ConsulAddr
	}
	if c.KVPath == nil {
		c.KVPath = o.KVPath
	}
	if c.ServiceName == nil {
		c.ServiceName = o.ServiceName
	}
	if c.TTLEndpoint == nil {
		c.TTLEndpoint = o.TTLEndpoint
	}
	if c.TTLID == nil {
		c.TTLID = o.TTLID
	}
	if c.EnableChecks == nil {
		c.EnableChecks = o.EnableChecks
	}
	if c.DebugMode == nil {
		c.DebugMode = o.DebugMode
	}
	if c.ToWatch == nil {
		c.ToWatch = o.ToWatch
	}
	return c
}

func defaultConfig() *serverConfig {
	return &serverConfig{
		Language:     StringPtr("english"),
		ConsulAddr:   StringPtr(fmt.Sprintf("http://%s:8500", os.Getenv("HOST_IP"))),
		KVPath:       StringPtr("/v1/kv/service/hello/"),
		ServiceName:  StringPtr("hello-http/"),
		TTLEndpoint:  StringPtr("/v1/agent/check/pass/"),
		TTLID:        StringPtr("hello-ttl"),
		EnableChecks: BoolPtr(true),
		DebugMode:    BoolPtr(false),
		ToWatch:      SlicePtr([]string{"hello-http/enable_checks"}),
	}
}

func loadConfig(filename string) (*serverConfig, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %v", filename, err)
	}
	defer f.Close()

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read '%s': %v", filename, err)
	}

	var cfg serverConfig
	json.Unmarshal(body, &cfg)

	return &cfg, nil
}

// BoolPtr returns a pointer to the given bool.
func BoolPtr(b bool) *bool {
	return &b
}

// BoolVal returns the value of the boolean at the pointer, or false if the
// pointer is nil.
func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// StringVal returns the value of the string at the pointer, or "" if the
// pointer is nil.
func StringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SlicePtr returns a pointer to the given string slice.
func SlicePtr(s []string) *[]string {
	return &s
}

// SliceVal returns the value of the slice at the pointer, or an empty
// slice if the pointer is nil
func SliceVal(s *[]string) []string {
	if s == nil {
		return []string{}
	}
	return *s
}
