package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	Listen              string `long:"listen" description:"Listen address" value-name:"[ADDR]:PORT" default:":49944"`
	NodeWsEndpoint      string `long:"ws-endpoint" description:"Node websocket endpoint" value-name:"<ws|wss>://ADDR[:PORT]" default:"ws://127.0.0.1:9944"`
	ProbeTimeoutSeconds uint32 `short:"t" long:"timeout" description:"Probe timeout in seconds" value-name:"n" default:"1"`
	LogLevel            uint32 `long:"log-level" description:"The log level (0 ~ 6)" value-name:"n" default:"4"`
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

type ProbeRequest struct {
	Name    string
	Request []byte
}

var probeRequests []ProbeRequest

func init() {
	probeRequests = []ProbeRequest{
		{
			Name:    "system_health",
			Request: rpc.SystemHealth(0),
		},
		{
			Name:    "system_health",
			Request: rpc.SystemChain(0),
		},
		{
			Name:    "system_properties",
			Request: rpc.SystemProperties(0),
		},
		{
			Name:    "chain_getBlockHash",
			Request: rpc.ChainGetBlockHash(0, 0),
		},
	}
}

func sendWsRequest(conn *ws.Conn, data []byte) (*rpc.JsonRpcResult, error) {
	v := &rpc.JsonRpcResult{}

	if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
		return nil, fmt.Errorf("conn.WriteMessage: %w", err)
	}

	if err := conn.ReadJSON(v); err != nil {
		return nil, fmt.Errorf("conn.ReadJSON: %w", err)
	}

	if v.Error != nil {
		return nil, fmt.Errorf("Websocket returned an error: %s", v.Error.Message)
	}

	return v, nil
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	fmt.Printf("Substrate Node Livness Probe %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)

	log.SetLevel(log.Level(opts.LogLevel))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Debugf("Received probe request from %s %s", r.RemoteAddr, r.Header.Get("User-Agent"))

		dialer := ws.Dialer{
			HandshakeTimeout: time.Duration(opts.ProbeTimeoutSeconds) * time.Second,
		}

		conn, _, err := dialer.Dial(opts.NodeWsEndpoint, nil)

		if conn != nil {
			defer conn.Close()
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err = fmt.Errorf("Dial: %w", err)
			log.WithError(err).Error()
			return
		}

		for _, p := range probeRequests {
			if r, err := sendWsRequest(conn, p.Request); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.WithError(err).Error()
				return
			} else {
				log.Debugf("RPC %s result: %s", p.Name, r.Result)
			}
		}

		elapsed := time.Since(start)
		log.Infof("Probe succeeded, time elapsed %s", elapsed)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Infof("Serving requests to /healthz on %s", opts.Listen)
	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
