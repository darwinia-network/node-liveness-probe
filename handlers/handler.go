package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Prober interface {
	Probe(*ws.Conn) (int, error)
}

type ProbeHandler struct {
	Prober

	WsEndpoint string
}

func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	start := time.Now()
	log.Debugf("Received request %s from %s %s", r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

	timeout, err := h.parseTimeoutFromQuery(r)
	if err != nil {
		statusCode = http.StatusInternalServerError
		err = fmt.Errorf("parseTimeoutFromQuery: %w", err)
		log.Error(err)
	} else if statusCode, err = h.dialAndProbe(timeout); err != nil {
		log.Warn(err)
	}

	elapsed := time.Since(start)
	log.Infof("Probe %s returning %d in %s", r.URL.Path, statusCode, elapsed)

	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
}

func (h *ProbeHandler) dialAndProbe(wsHandshakeTimeout *time.Duration) (int, error) {
	dialer := &ws.Dialer{
		HandshakeTimeout: *wsHandshakeTimeout,
	}

	log.Debugf("Dialer: %+v", dialer)

	conn, _, err := dialer.Dial(h.WsEndpoint, nil)

	if conn != nil {
		defer conn.Close()
	}

	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("Dial: %w", err)
	} else {
		return h.Prober.Probe(conn)
	}
}

func (h *ProbeHandler) parseTimeoutFromQuery(r *http.Request) (*time.Duration, error) {
	var (
		timeoutInSecond int
		err             error
	)

	if t := r.URL.Query().Get("timeout"); t == "" {
		timeoutInSecond = 1
	} else if timeoutInSecond, err = strconv.Atoi(t); err != nil {
		return nil, err
	}

	timeout := time.Duration(timeoutInSecond) * time.Second

	return &timeout, nil
}
