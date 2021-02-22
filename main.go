package main

import (
	"flag"
	"net/http"

	"github.com/darwinia-network/node-liveness-probe/handlers"
	"github.com/darwinia-network/node-liveness-probe/probes"
	"k8s.io/klog/v2"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var (
	listen                = flag.String("listen", ":49944", "Listen address")
	wsEndpoint            = flag.String("ws-endpoint", "ws://127.0.0.1:9944", "Substrate node WebSocket endpoint")
	blockThresholdSeconds = flag.Float64("block-threshold-seconds", 300, "/healthz_block returns unhealthy if node's latest block is older than threshold")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	klog.Infof("Substrate Node Livness Probe %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)

	http.Handle("/healthz", &handlers.ProbeHandler{
		Prober:     &probes.LivenessProbe{},
		WsEndpoint: *wsEndpoint,
	})
	http.Handle("/healthz_block", &handlers.ProbeHandler{
		Prober:     &probes.LivenessBlockProbe{BlockThresholdSeconds: *blockThresholdSeconds},
		WsEndpoint: *wsEndpoint,
	})
	http.Handle("/readiness", &handlers.ProbeHandler{
		Prober:     &probes.ReadinessProbe{},
		WsEndpoint: *wsEndpoint,
	})

	klog.Infof("Serving requests on %s", *listen)
	err := http.ListenAndServe(*listen, nil)
	if err != nil {
		klog.Fatalf("failed to start http server with error: %v", err)
	}
}
