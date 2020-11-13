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
			Name:    "system_chain",
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

func (p *LivenessProbe) Probe(conn *ws.Conn) (error, int) {
	for _, p := range livenessProbeRequests {
		if _, err := sendWsRequest(conn, p.Name, p.Request); err != nil {
			return err, http.StatusInternalServerError
		}
	}

	return nil, http.StatusOK
}

func sendWsRequest(conn *ws.Conn, name string, data []byte) (*rpc.JsonRpcResult, error) {
	v := &rpc.JsonRpcResult{}

	if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
		return nil, fmt.Errorf("conn.WriteMessage: %w", err)
	}

	if err := conn.ReadJSON(v); err != nil {
		return nil, fmt.Errorf("conn.ReadJSON: %w", err)
	}

	if v.Error != nil {
		return nil, fmt.Errorf("RPC %s error: %s", name, v.Error.Message)
	}

	log.Debugf("RPC %s result: %+v", name, v.Result)
	return v, nil
}
