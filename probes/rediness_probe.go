package probes

import (
	"fmt"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
	log "github.com/sirupsen/logrus"
)

type ReadinessProbe struct{}

func (p ReadinessProbe) Probe(conn *ws.Conn) (error, int) {
	if ready, err := isNodeReady(conn); err != nil {
		return err, http.StatusInternalServerError
	} else if !*ready {
		return fmt.Errorf("Node is not ready"), http.StatusServiceUnavailable
	}

	return nil, http.StatusOK
}

func isNodeReady(conn *ws.Conn) (*bool, error) {
	// Should use `system_syncState` once the PR
	// https://github.com/paritytech/substrate/pull/7315 is available.
	if r, err := sendWsRequest(conn, rpc.SystemHealth(0)); err != nil {
		return nil, err
	} else {
		log.Debugf("RPC system_health result: %+v", r)

		if h := r.ToSysHealth(); h == nil {
			return nil, fmt.Errorf("ToSysHealth returned nil, rpc result: %+v", *r)
		} else {
			r := !h.IsSyncing && h.Peers > 0
			return &r, nil
		}
	}
}
