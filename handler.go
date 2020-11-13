package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Prober interface {
	Probe(*ws.Conn) (error, int)
}

type ProbeHandler struct {
	Prober

	wsHandshakeTimeout time.Duration
}

func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	start := time.Now()
	log.Debugf("Received request %s from %s %s", r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

	err := h.parseTimeoutFromQuery(r)
	if err != nil {
		statusCode = http.StatusInternalServerError
		err = fmt.Errorf("parseTimeoutFromQuery: %w", err)
		log.Error(err)
	} else if statusCode, err = h.dialAndProbe(); err != nil {
		log.Warn(err)
	}

	elapsed := time.Since(start)
	log.Infof("Probe %s returning %d in %s", r.URL.Path, statusCode, elapsed)

	w.WriteHeader(statusCode)
	if statusCode == http.StatusOK {
		w.Write([]byte("OK"))
	}
}

func (h *ProbeHandler) dialAndProbe() (int, error) {
	dialer := &ws.Dialer{
		HandshakeTimeout: h.wsHandshakeTimeout,
	}

	log.Debugf("Dialer: %+v", dialer)

	conn, _, err := dialer.Dial(opts.NodeWsEndpoint, nil)

	if conn != nil {
		defer conn.Close()
	}

	var statusCode int

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Dial: %w", err)
	} else {
		err, statusCode = h.Prober.Probe(conn)
		return statusCode, err
	}
}

func (h *ProbeHandler) parseTimeoutFromQuery(r *http.Request) error {
	var (
		timeoutInSecond int
		err             error
	)

	if t := r.URL.Query().Get("timeout"); t == "" {
		timeoutInSecond = 1
	} else if timeoutInSecond, err = strconv.Atoi(t); err != nil {
		return err
	}

	h.wsHandshakeTimeout = time.Duration(timeoutInSecond) * time.Second

	return nil
}
