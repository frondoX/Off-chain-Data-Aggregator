package uniswap_v2

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"off-chain-data-aggregator/api"
	"off-chain-data-aggregator/worker"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func init() {
	worker.RegisterIndexer("uniswap_v2", func() api.ContractIndexer {
		return &Indexer{}
	})
}

type Indexer struct {
	swapTopic string
	swapEvent abi.Event
}

func (i *Indexer) Name() string {
	return "uniswap_v2"
}

func (i *Indexer) Init(ctx context.Context, cfg *api.RuntimeConfig) error {
	i.swapTopic = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"

	abiJson := `[
	{
		"anonymous": true,
		"inputs": [
			{"indexed":true,"name":"sender","type":"address"},
			{"indexed":false,"name":"amount0In","type":"uint256"},
			{"indexed":false,"name":"amount1In","type":"uint256"},
			{"indexed":false,"name":"amount0Out","type":"uint256"},
			{"indexed":false,"name":"amount1Out","type":"uint256"},
			{"indexed":true,"name":"to","type":"address"}
		],
		"name":"Swap",
		"type":"event"
	}
]`
	parsedAbi, err := abi.JSON(bytes.NewBufferString(abiJson))
	if err != nil {
		return err
	}
	evt, ok := parsedAbi.Events["Swap"]
	if !ok {
		return fmt.Errorf("swap event not found in abi")
	}
	i.swapEvent = evt
	return nil
}

func (i *Indexer) Filter() []string {
	return []string{i.swapTopic}
}

func (i *Indexer) OnLog(ctx context.Context, log *api.LogEvent) ([]*api.Record, error) {
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("topics length invalid, got %d", len(log.Topics))
	}

	sender := common.HexToAddress(log.Topics[1])
	to := common.HexToAddress(log.Topics[2])

	// 只解非indexed的四个参数：amount0In,amount1In,amount0Out,amount1Out
	args, err := i.swapEvent.Inputs.Unpack(log.Data)
	if err != nil {
		return nil, fmt.Errorf("unpack data err: %w", err)
	}

	rec := &api.Record{
		Table: "uni_v2_swap",
		Data: map[string]any{
			"block_number": log.BlockNumber,
			"tx_hash":      log.TxHash,
			"log_index":    log.LogIndex,
			"pair":         log.Address,
			"sender":       sender.Hex(),
			"to":           to.Hex(),
			"amount0In":    args[0].(*big.Int).String(),
			"amount1In":    args[1].(*big.Int).String(),
			"amount0Out":   args[2].(*big.Int).String(),
			"amount1Out":   args[3].(*big.Int).String(),
			"raw_data":     hex.EncodeToString(log.Data),
		},
	}
	return []*api.Record{rec}, nil
}
