package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/darwinia-network/node-liveness-probe/handlers"
	"github.com/darwinia-network/node-liveness-probe/probes"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	Listen                string  `long:"listen" description:"Listen address" value-name:"[ADDR]:PORT" default:":49944"`
	NodeWsEndpoint        string  `long:"ws-endpoint" description:"Node websocket endpoint" value-name:"<ws|wss>://ADDR[:PORT]" default:"ws://127.0.0.1:9944"`
	LogLevel              uint32  `long:"log-level" description:"The log level (0 ~ 6), use 5 for debugging, see https://pkg.go.dev/github.com/sirupsen/logrus#Level" value-name:"N" default:"4"`
	BlockThresholdSeconds float64 `long:"block-threshold-seconds" description:"/healthz_block returns unhealthy if node's latest block is older than threshold" value-name:"N" default:"120"`
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	fmt.Printf("Substrate Node Livness Probe %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)

	log.SetLevel(log.Level(opts.LogLevel))

	http.Handle("/healthz", &handlers.ProbeHandler{
		Prober:     &probes.LivenessProbe{},
		WsEndpoint: opts.NodeWsEndpoint,
	})
	http.Handle("/healthz_block", &handlers.ProbeHandler{
		Prober:     &probes.LivenessBlockProbe{BlockThresholdSeconds: opts.BlockThresholdSeconds},
		WsEndpoint: opts.NodeWsEndpoint,
	})
	http.Handle("/readiness", &handlers.ProbeHandler{
		Prober:     &probes.ReadinessProbe{},
		WsEndpoint: opts.NodeWsEndpoint,
	})

	log.Infof("Serving requests on %s", opts.Listen)
	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
