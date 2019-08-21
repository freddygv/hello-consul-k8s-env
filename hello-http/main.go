package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/matryer/way"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	limiterRate    = 0.1
	limiterBurst   = 2
	gRPCAddr       = "localhost:9090"
	prometheusAddr = "localhost:9091"
	defaultAddr    = "localhost:8080"
	defaultCfg     = "config.json"
)

var (
	httpReqs = prometheus.NewCounter(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Count of HTTP requests processed.",
    })
)

func init() {
	prometheus.MustRegister(httpReqs)
}

func main() {
	var (
		httpAddr   = flag.String("addr", defaultAddr, "Hello service address.")
		configFile = flag.String("cfg-file", defaultCfg, "Path to config file.")
	)
	flag.Parse()

	log.Printf("[INFO] Starting server...")

	s := newServer(*configFile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, key := range SliceVal(s.cfg.ToWatch) {
		log.Printf("[INFO] Running watch for key '%s'", key)
		go s.watchKV(ctx, key, limiterRate, limiterBurst)
	}

	go s.captureReload(ctx, StringVal(configFile))

	log.Printf("[INFO] gRPC health check listening on '%s'...", gRPCAddr)
	go s.runGRPC(ctx, gRPCAddr)

	log.Printf("[INFO] Exposing Prometheus metrics on '%s'...", prometheusAddr)
	go s.runPrometheus(prometheusAddr)

	log.Printf("[INFO] Hello service with HTTP check listening on %s", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, s.router))
}

type server struct {
	router *way.Router
	cfg    *serverConfig
}

func newServer(cfgFile string) *server {
	config, err := loadConfig(cfgFile)
	if err != nil {
		log.Printf("[WARN] failed to load config from file '%s', using default. err: %v", cfgFile, err)
	}
	config = config.merge(defaultConfig())

	s := server{
		router: way.NewRouter(),
		cfg:    config,
	}

	s.router.HandleFunc("GET", "/hello", s.handleHello())
	s.router.HandleFunc("GET", "/healthz", s.handleHealth())
	s.router.HandleFunc("PUT", "/health/pass", s.enableHealth())
	s.router.HandleFunc("PUT", "/health/fail", s.disableHealth())

	return &s
}

// Reload config from file on HUP
func (s *server) captureReload(ctx context.Context, cfgFile string) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)

	for {
		select {
		case sig := <-sigCh:
			log.Printf("[INFO] captured signal: %v. reloading config...", sig)
			config, err := loadConfig(cfgFile)
			if err != nil {
				log.Printf("[WARN] failed to load config from file '%s', using default. err: %v", cfgFile, err)
			}
			s.cfg.mu.Lock()
			{
				s.cfg = config.merge(s.cfg)
			}
			s.cfg.mu.Unlock()
		}
	}
}

func (s *server) runPrometheus(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}

// Run a gRPC server exclusively for health checking
func (s *server) runGRPC(ctx context.Context, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[ERR] grpc health: %v", addr, err)
	}

	gs := grpc.NewServer()
	server := health.NewServer()
	grpc_health_v1.RegisterHealthServer(gs, server)

	svcName := strings.TrimSuffix(StringVal(s.cfg.ServiceName), "/")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var enableChecks bool
				s.cfg.mu.RLock()
				{
					enableChecks = BoolVal(s.cfg.EnableChecks)
				}
				s.cfg.mu.RUnlock()

				switch enableChecks {
				case true:
					server.SetServingStatus(svcName, grpc_health_v1.HealthCheckResponse_SERVING)
				case false:
					server.SetServingStatus(svcName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("[ERR] grpc health: failed to serve: %v", err)
	}
}

func (s *server) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.cfg.mu.RLock()
		defer s.cfg.mu.RUnlock()

		switch StringVal(s.cfg.Language) {
		case "french":
			fmt.Fprintln(w, "Bonjour Monde")
		case "portuguese":
			fmt.Fprintln(w, "Olá Mundo")
		case "spanish":
			fmt.Fprintln(w, "Hola Mundo")
		default:
			fmt.Fprintln(w, "Hello World")
		}
		httpReqs.Inc()
	}
}

func (s *server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.cfg.mu.RLock()
		defer s.cfg.mu.RUnlock()

		// Fail check if checks aren't enabled
		if !BoolVal(s.cfg.EnableChecks) {
			w.WriteHeader(http.StatusGone)
			return
		}

		fmt.Fprintln(w, "I'm alive")
		httpReqs.Inc()
	}
}

func (s *server) disableHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.cfg.mu.Lock()
		defer s.cfg.mu.Unlock()

		s.cfg.EnableChecks = BoolPtr(false)
		fmt.Fprintln(w, "Health endpoint disabled.")
		httpReqs.Inc()
	}
}

func (s *server) enableHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.cfg.mu.Lock()
		defer s.cfg.mu.Unlock()

		s.cfg.EnableChecks = BoolPtr(true)
		fmt.Fprintln(w, "Health endpoint enabled.")
		httpReqs.Inc()
	}
}

// watchKV watches a Key/Value pair in Consul for changes and sets the value internally
// See below for implementation details:
// https://www.consul.io/api/features/blocking.html#implementation-details
func (s *server) watchKV(ctx context.Context, key string, limit rate.Limit, burst int) {
	var index uint64 = 1
	var lastIndex uint64

	limiter := rate.NewLimiter(limit, burst)

	for {
		// Wait until limiter allows request to happen
		if err := limiter.Wait(context.Background()); err != nil {
			log.Printf("[ERR] watch '%s': failed to wait for limiter", key)
			continue
		}

		// Make blocking query to watch key
		target := fmt.Sprintf("%s%s%s?index=%d", StringVal(s.cfg.ConsulAddr), StringVal(s.cfg.KVPath), key, index)
		resp, err := http.Get(target)
		if err != nil {
			log.Printf("[ERR] watch '%s': failed to get '%s': %v", key, target, err)
			continue
		}

		// Parse the raft index for this key (X-Consul-Index)
		header := resp.Header
		indexStr := header.Get("X-Consul-Index")
		if indexStr != "" {
			index, err = strconv.ParseUint(indexStr, 10, 64)
			if err != nil {
				log.Printf("[ERR] watch '%s': failed to parse X-Consul-Index: %v", key, err)
				continue
			}
		}
		// Reset if it goes backwards or is 0
		// See: https://www.consul.io/api/features/blocking.html#implementation-details
		if index < lastIndex || index == 0 {
			index = 1
			lastIndex = 1

			// TODO: Continuing implies we don't trust the data on the server
			continue
		}
		lastIndex = index

		data := make([]keyResponse, 0)
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()

		// Key might not exist yet
		if len(data) == 0 {
			log.Printf("[WARN] watch '%s': empty response, key does not exist", key)
			continue
		}

		// We are not recursing on a key-prefix so these arrays will only return one value
		decoded, err := base64.StdEncoding.DecodeString(data[0].Value)
		if err != nil {
			log.Printf("[ERR] watch '%s': failed to decode value: '%s'", key, data[0].Value)
			continue
		}
		strVal := string(decoded)

		err = nil
		switch key {
		case "language":
			s.setLanguage(strVal)
		case StringVal(s.cfg.ServiceName) + "enable_checks":
			err = s.setEnableChecks(strVal)
		}
		if err != nil {
			log.Printf("[ERR] watch '%s': %v", key, err)
			continue
		}

		log.Printf("[INFO] watch '%s': updated to %s", key, strVal)
	}
}

func (s *server) setLanguage(lang string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()

	s.cfg.Language = StringPtr(lang)
}

func (s *server) setEnableChecks(val string) error {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()

	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("failed to parse enable_checks bool '%s': %v", val, err)
	}
	s.cfg.EnableChecks = BoolPtr(parsed)
	return nil
}

type keyResponse struct {
	LockIndex   uint64
	Key         string
	Flags       int
	Value       string
	CreateIndex uint64
	ModifyIndex uint64
}
