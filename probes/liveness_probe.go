package probes

import (
	"fmt"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
	"k8s.io/klog/v2"
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

func (p *LivenessProbe) Probe(conn *ws.Conn) (int, error) {
	for _, p := range livenessProbeRequests {
		if _, err := sendWsRequest(conn, p.Name, p.Request); err != nil {
			return http.StatusServiceUnavailable, err
		}
	}

	return http.StatusOK, nil
}

func sendWsRequest(conn *ws.Conn, name string, data []byte) (*rpc.JsonRpcResult, error) {
	klog.V(5).Infof("sendWsRequest: %s", data)
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

	klog.V(4).Infof("RPC %s result: %+v", name, v.Result)
	return v, nil
}
