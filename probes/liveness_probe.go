package probes

import (
	"fmt"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
	log "github.com/sirupsen/logrus"
)

type LivenessProbe struct{}

type ProbeRequest struct {
	Name    string
	Request []byte
}

var livenessProbeRequests []ProbeRequest

func init() {
	livenessProbeRequests = []ProbeRequest{
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

func (p LivenessProbe) Probe(conn *ws.Conn) (error, int) {
	for _, p := range livenessProbeRequests {
		if r, err := sendWsRequest(conn, p.Request); err != nil {
			return err, http.StatusInternalServerError
		} else {
			log.Debugf("RPC %s result: %+v", p.Name, r.Result)
		}
	}

	return nil, http.StatusOK
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
		return nil, fmt.Errorf("RPC error: %s", v.Error.Message)
	}

	return v, nil
}
