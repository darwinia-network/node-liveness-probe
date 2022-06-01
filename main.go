package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/darwinia-network/node-liveness-probe/handlers"
	"github.com/darwinia-network/node-liveness-probe/probes"
	"k8s.io/klog/v2"
)

type stringListValue []string

func (i *stringListValue) String() string {
	return strings.Join(*i, ",")
}
func (i *stringListValue) Set(s string) error {
	*i = append(*i, s)
	return nil
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"

	wsEndpoints           stringListValue
	listen                = flag.String("listen", ":49944", "Listen address")
	blockThresholdSeconds = flag.Float64("block-threshold-seconds", 300, "/healthz_block returns unhealthy if node's latest block is older than threshold")
)

func initFlags() {
	klog.InitFlags(nil)
	flag.Var(&wsEndpoints, "ws-endpoint", "Substrate node WebSocket endpoint; may be specified multiple times to probe both relaychain and parachain sequentially (default \"ws://127.0.0.1:9944\")")
	flag.Set("logtostderr", "true")
	flag.Parse()
	if len(wsEndpoints) == 0 {
		wsEndpoints = append(wsEndpoints, "ws://127.0.0.1:9944")
	}
}

func main() {
	initFlags()
	klog.Infof("Substrate Node Livness Probe %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)

	http.Handle("/healthz", &handlers.ProbeHandler{
		Prober:      &probes.LivenessProbe{},
		WsEndpoints: wsEndpoints,
	})
	http.Handle("/healthz_block", &handlers.ProbeHandler{
		Prober:      &probes.LivenessBlockProbe{BlockThresholdSeconds: *blockThresholdSeconds},
		WsEndpoints: wsEndpoints,
	})
	http.Handle("/readiness", &handlers.ProbeHandler{
		Prober:      &probes.ReadinessProbe{},
		WsEndpoints: wsEndpoints,
	})

	klog.Infof("Serving requests on %s", *listen)
	err := http.ListenAndServe(*listen, nil)
	if err != nil {
		klog.Fatalf("failed to start http server with error: %v", err)
	}
}
