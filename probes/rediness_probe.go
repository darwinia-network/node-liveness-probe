package probes

import (
	"fmt"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
)

type ReadinessProbe struct{}

func (p *ReadinessProbe) Probe(conn *ws.Conn) (int, error) {
	if ready, err := isNodeReady(conn); err != nil {
		return http.StatusServiceUnavailable, err
	} else if !*ready {
		return http.StatusServiceUnavailable, fmt.Errorf("Node is not ready")
	}

	return http.StatusOK, nil
}

func isNodeReady(conn *ws.Conn) (*bool, error) {
	// Should use `system_syncState` once the PR
	// https://github.com/paritytech/substrate/pull/7315 is available.
	if r, err := sendWsRequest(conn, "system_health", rpc.SystemHealth(0)); err != nil {
		return nil, err
	} else if h := r.ToSysHealth(); h == nil {
		return nil, fmt.Errorf("ToSysHealth returned nil, RPC response: %+v", *r)
	} else {
		r := !h.IsSyncing && h.Peers > 0
		return &r, nil
	}
}
