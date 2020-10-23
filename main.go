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
	Listen         string `short:"l" long:"listen" description:"Listen address" value-name:"[ADDR]:PORT" default:":49944"`
	NodeWsEndpoint string `short:"e" long:"endpoint" description:"Node websocket endpoint" value-name:"<ws|wss>://ADDR[:PORT]" default:"ws://127.0.0.1:9944"`
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func sendWsRequest(conn *ws.Conn, w http.ResponseWriter, data []byte) *rpc.JsonRpcResult {
	v := &rpc.JsonRpcResult{}

	if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = fmt.Errorf("conn.WriteMessage: %w", err)
		log.WithError(err).Error()
		return nil
	}

	if err := conn.ReadJSON(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = fmt.Errorf("conn.ReadJSON: %w", err)
		log.WithError(err).Error()
		return nil
	}

	if v.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := fmt.Errorf("Websocket returned an error: %s", v.Error.Message)
		log.WithError(err).Error()
		return nil
	}

	return v
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	fmt.Printf("Substrate Node Livness Probe %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		var conn *ws.Conn
		var err error

		dialer := ws.Dialer{HandshakeTimeout: 2 * time.Second}
		if conn, _, err = dialer.Dial(opts.NodeWsEndpoint, nil); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err = fmt.Errorf("Dial: %w", err)
			log.WithError(err).Error()
			return
		}

		v := sendWsRequest(conn, w, rpc.SystemHealth(0))
		log.Infof("RPC call system_health succeeded: %s", v.Result)

		v = sendWsRequest(conn, w, rpc.SystemChain(0))
		log.Infof("RPC call system_chain succeeded: %s", v.Result)

		v = sendWsRequest(conn, w, rpc.SystemProperties(0))
		log.Infof("RPC call system_properties succeeded: %s", v.Result)

		v = sendWsRequest(conn, w, rpc.ChainGetBlockHash(0, 0))
		log.Infof("RPC call chain_getBlockHash succeeded: %s", v.Result)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Infof("Serving requests to /healthz on %s", opts.Listen)
	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
