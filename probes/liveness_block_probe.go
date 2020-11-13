package probes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/itering/substrate-api-rpc/rpc"
	log "github.com/sirupsen/logrus"
)

type LivenessBlockProbe struct {
	LivenessProbe

	lastBlockNumber int64
	lastBlockTime   time.Time

	BlockThresholdSeconds float64
}

func (p *LivenessBlockProbe) Probe(conn *ws.Conn) (error, int) {
	livenessProbeErr, livenessProbeStatusCode := p.LivenessProbe.Probe(conn)
	if livenessProbeErr != nil {
		return livenessProbeErr, livenessProbeStatusCode
	}

	if err := p.UpdateLatestBlock(conn); err != nil {
		return err, http.StatusInternalServerError
	}

	sinceLastBlockSeconds := time.Since(p.lastBlockTime).Seconds()
	if sinceLastBlockSeconds > p.BlockThresholdSeconds {
		err := fmt.Errorf(
			"The last block %d was obtained %.2f second(s) ago, above the threshold %.2f",
			p.lastBlockNumber,
			sinceLastBlockSeconds,
			p.BlockThresholdSeconds,
		)
		return err, http.StatusInternalServerError
	} else {
		log.Debugf(
			"The last block %d was obtained %.2f second(s) ago, below the threshold %.2f",
			p.lastBlockNumber,
			sinceLastBlockSeconds,
			p.BlockThresholdSeconds,
		)
	}

	// Inherit
	return livenessProbeErr, livenessProbeStatusCode
}

func rpcChainGetLatestBlock(id int) []byte {
	rpc := rpc.Param{Id: id, Method: "chain_getBlock", JsonRpc: "2.0"}
	b, _ := json.Marshal(rpc)
	return b
}

func (p *LivenessBlockProbe) UpdateLatestBlock(conn *ws.Conn) error {
	if r, err := sendWsRequest(conn, "chain_getBlock", rpcChainGetLatestBlock(0)); err != nil {
		return err
	} else if blk := r.ToBlock(); blk == nil {
		return fmt.Errorf("ToBlock returned nil, RPC response: %+v", *r)
	} else {
		return p.SetLatestBlock(blk)
	}
}

func (p *LivenessBlockProbe) SetLatestBlock(r *rpc.BlockResult) error {
	latestBlockNumberHex := r.Block.Header.Number
	latestBlockNumber, err := strconv.ParseInt(latestBlockNumberHex, 0, 64)

	if err != nil {
		return err
	}

	if latestBlockNumber != p.lastBlockNumber {
		p.lastBlockTime = time.Now()
		p.lastBlockNumber = latestBlockNumber
	}

	return nil
}
