package main

import (
	"fmt"
	"net/http"
	"time"

	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Prober interface {
	Probe(*ws.Conn) (error, int)
}

type ProbeHandler struct {
	Prober
}

func (h ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Debugf("Received request %s from %s %s", r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

	dialer := ws.Dialer{
		HandshakeTimeout: time.Duration(opts.ProbeTimeoutSeconds) * time.Second,
	}

	conn, _, err := dialer.Dial(opts.NodeWsEndpoint, nil)

	if conn != nil {
		defer conn.Close()
	}

	var statusCode int

	if err != nil {
		statusCode = http.StatusInternalServerError
		err = fmt.Errorf("Dial: %w", err)
		log.Warn(err)
	} else if err, statusCode = h.Prober.Probe(conn); err != nil {
		log.Warn(err)
	}

	elapsed := time.Since(start)
	log.Infof("Probe done, time elapsed %s", elapsed)

	w.WriteHeader(statusCode)
	if statusCode == http.StatusOK {
		w.Write([]byte("OK"))
	}
}
