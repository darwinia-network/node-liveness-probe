package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	ws "github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

type Prober interface {
	Probe(*ws.Conn) (int, error)
}

type ProbeHandler struct {
	Prober

	WsEndpoints []string
}

func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	start := time.Now()
	klog.V(4).Infof("Received request %s from %s %s", r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

	timeout, err := h.parseTimeoutFromQuery(r)
	if err != nil {
		statusCode = http.StatusInternalServerError
		err = fmt.Errorf("parseTimeoutFromQuery: %w", err)
		klog.Error(err)
	} else if statusCode, err = h.dialAndProbeAll(*timeout); err != nil {
		klog.Warning(err)
	}

	elapsed := time.Since(start)
	klog.Infof("Probe %s returning %d in %s", r.URL.Path, statusCode, elapsed)

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
}

func (h *ProbeHandler) dialAndProbeAll(timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)

	var (
		statusCode int
		err        error
	)

	for _, ep := range h.WsEndpoints {
		if statusCode, err = h.dialAndProbe(ep, timeout, deadline); err != nil {
			return statusCode, err
		}
	}

	return statusCode, nil
}

func (h *ProbeHandler) dialAndProbe(endpoint string, timeout time.Duration, deadline time.Time) (int, error) {
	dialer := &ws.Dialer{
		HandshakeTimeout: timeout,
	}
	conn, _, err := dialer.Dial(endpoint, nil)

	if conn != nil {
		defer conn.Close()
	}

	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("Dial: %w", err)
	}

	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)

	return h.Prober.Probe(conn)
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
