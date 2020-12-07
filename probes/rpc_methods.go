package probes

import (
	"encoding/json"

	"github.com/itering/substrate-api-rpc/rpc"
)

func rpcChainGetLatestBlock(id int) []byte {
	rpc := rpc.Param{Id: id, Method: "chain_getBlock", JsonRpc: "2.0"}
	b, _ := json.Marshal(rpc)
	return b
}

func rpcChainGetFinalizedHead(id int) []byte {
	rpc := rpc.Param{Id: id, Method: "chain_getFinalizedHead", JsonRpc: "2.0"}
	b, _ := json.Marshal(rpc)
	return b
}
