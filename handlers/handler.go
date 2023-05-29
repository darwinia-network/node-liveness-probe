package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

	MetricsEndpoint                string
	UseMetrics                     bool
	FinalizedBlockThresholdSeconds int64
	metricsStartTime               time.Time
	metricsFinalizedBlockNumber    int64

	WsEndpoints []string
}

// metricsProbe probes the metrics endpoint of a liveness block
// Returns an error if the metrics probe fails or the finalized block number
// did not increase within FinalizedBlockThresholdSeconds
func (p *ProbeHandler) metricsProbe() (int, error) {
	resp, err := http.Get(p.MetricsEndpoint)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("metrics probe failed with status code %d", resp.StatusCode)
	}
	var (
		r                    = bufio.NewReader(resp.Body)
		commentSymbol        = []byte("#")
		finalizedSymbol      = []byte("status=\"finalized\"")
		bestSymbol           = []byte("status=\"best\"")
		splitSymbol          = []byte("\"} ")
		finalizedBlockNumber int64
		bestBlockNumber      int64
	)

	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}

		if bytes.HasPrefix(line, commentSymbol) {
			continue
		}
		data := bytes.Split(line, splitSymbol)
		if len(data) < 2 {
			continue
		}
		blockNumber := Bytes2Int64(bytes.TrimSpace(data[len(data)-1]))
		if blockNumber <= 0 {
			continue
		}
		if bytes.Contains(line, finalizedSymbol) {
			finalizedBlockNumber = blockNumber
			continue
		}
		if bytes.Contains(line, bestSymbol) {
			bestBlockNumber = blockNumber
			continue
		}
	}
	now := time.Now()
	if p.metricsStartTime.IsZero() {
		p.metricsStartTime = now
	}
	klog.Infof("Retrieved block, now best: #%d, now finalized: #%d, last finalized: #%d, last fetch time is %s", bestBlockNumber, finalizedBlockNumber, p.metricsFinalizedBlockNumber, p.metricsStartTime)
	if finalizedBlockNumber <= 0 || bestBlockNumber <= 0 {
		p.metricsStartTime = now
		return http.StatusOK, nil
	}

	if finalizedBlockNumber == p.metricsFinalizedBlockNumber && now.Sub(p.metricsStartTime).Seconds() >= float64(p.FinalizedBlockThresholdSeconds) {
		return http.StatusServiceUnavailable, fmt.Errorf("finalized did not increase within %dS", p.FinalizedBlockThresholdSeconds)
	}

	if finalizedBlockNumber != p.metricsFinalizedBlockNumber {
		p.metricsFinalizedBlockNumber = finalizedBlockNumber
		p.metricsStartTime = now
	}
	return resp.StatusCode, nil
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
	if statusCode == http.StatusOK && h.UseMetrics {
		if statusCode, err = h.metricsProbe(); err != nil {
			klog.Warning(err)
		}
		klog.Infof("Metrics probe %s returning %d in %s", r.URL.Path, statusCode, elapsed)
	}
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

func Bytes2Int64(b []byte) int64 {
	result, _ := strconv.Atoi(string(b))
	return int64(result)
}
