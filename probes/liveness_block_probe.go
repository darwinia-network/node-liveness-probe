package probes

import (
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

	bestBlock      Block
	finalizedBlock Block

	BlockThresholdSeconds float64
}

type Block struct {
	Number    int64
	UpdatedAt time.Time
}

func (b *Block) IsStale(thresholdSeconds float64, status string) error {
	sinceLastBlockSeconds := time.Since(b.UpdatedAt).Seconds()

	if sinceLastBlockSeconds > thresholdSeconds {
		return fmt.Errorf(
			"The %s block %d was obtained %.2f second(s) ago, above the threshold %.2f",
			status,
			b.Number,
			sinceLastBlockSeconds,
			thresholdSeconds,
		)
	} else {
		log.Debugf(
			"The %s block %d was obtained %.2f second(s) ago, below the threshold %.2f",
			status,
			b.Number,
			sinceLastBlockSeconds,
			thresholdSeconds,
		)
		return nil
	}
}

func (p *LivenessBlockProbe) Probe(conn *ws.Conn) (int, error) {
	livenessProbeStatusCode, livenessProbeErr := p.LivenessProbe.Probe(conn)
	if livenessProbeErr != nil {
		return livenessProbeStatusCode, livenessProbeErr
	}

	if err := p.UpdateBlock(conn); err != nil {
		return http.StatusServiceUnavailable, err
	}

	log.Infof("Retrieved block, best: #%d, finalized: #%d", p.bestBlock.Number, p.finalizedBlock.Number)

	errBestBlock := p.bestBlock.IsStale(p.BlockThresholdSeconds, "best")
	errFinalizedBlock := p.finalizedBlock.IsStale(p.BlockThresholdSeconds, "finalized")

	if errBestBlock != nil {
		return http.StatusServiceUnavailable, errBestBlock
	} else if errFinalizedBlock != nil {
		return http.StatusServiceUnavailable, errFinalizedBlock
	}

	// Inherit
	return livenessProbeStatusCode, livenessProbeErr
}

func (p *LivenessBlockProbe) UpdateBlock(conn *ws.Conn) error {
	// Best block
	if r, err := sendWsRequest(conn, "chain_getBlock", rpcChainGetLatestBlock(0)); err != nil {
		return err
	} else if err := setBlock(r, &p.bestBlock); err != nil {
		return err
	} else
	// Finalized block
	if r, err := sendWsRequest(conn, "chain_getFinalizedHead", rpcChainGetFinalizedHead(0)); err != nil {
		return err
	} else if r, err := sendWsRequest(conn, "chain_getBlock", rpc.ChainGetBlock(0, r.Result.(string))); err != nil {
		return err
	} else {
		return setBlock(r, &p.finalizedBlock)
	}
}

func setBlock(r *rpc.JsonRpcResult, b *Block) error {
	var blk *rpc.BlockResult
	if blk = r.ToBlock(); blk == nil {
		return fmt.Errorf("ToBlock returned nil, RPC response: %+v", *r)
	}

	blkNumberHex := blk.Block.Header.Number
	blkNumber, err := strconv.ParseInt(blkNumberHex, 0, 64)

	if err != nil {
		return err
	}

	if blkNumber != b.Number {
		b.UpdatedAt = time.Now()
		b.Number = blkNumber
	}

	return nil
}
