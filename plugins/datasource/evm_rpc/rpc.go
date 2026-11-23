package evm_rpc

import (
	"context"
	"fmt"
	"math/big"
	"off-chain-data-aggregator/api"
	"off-chain-data-aggregator/worker"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func init() {
	worker.RegisterDataSource("evm_rpc", func() api.DataSource {
		return &EVMRPC{}
	})
}

type EVMRPC struct {
	client      *ethclient.Client
	rpcEndpoint string
	startHeight uint64
	ch          chan *api.ChainEvent
}

func (e *EVMRPC) Name() string {
	return "evm_rpc"
}

// Init 从传入的顶层配置拿rpc和起始高度
func (e *EVMRPC) Init(ctx context.Context, cfg *api.RuntimeConfig) error {
	e.rpcEndpoint = cfg.Chain.RPCEndpoint
	e.startHeight = cfg.Chain.StartHeight
	client, err := ethclient.Dial(e.rpcEndpoint)
	if err != nil {
		return err
	}
	e.client = client
	e.ch = make(chan *api.ChainEvent, 8)
	go e.pollLoop(ctx)
	return nil
}

func (e *EVMRPC) Subscribe(ctx context.Context) (<-chan *api.ChainEvent, error) {
	return e.ch, nil
}

func (e *EVMRPC) pollLoop(ctx context.Context) {
	currentHeight := e.startHeight
	swapTopic := common.HexToHash("0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822")

	for {
		select {
		case <-ctx.Done():
			close(e.ch)
			e.client.Close()
			return
		default:
		}

		latest, err := e.client.BlockNumber(ctx)
		if err != nil {
			fmt.Printf("get latest block err: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if currentHeight > latest {
			time.Sleep(2 * time.Second)
			continue
		}

		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(currentHeight)),
			ToBlock:   big.NewInt(int64(currentHeight)),
			Topics: [][]common.Hash{
				{swapTopic},
			},
		}

		logs, err := e.client.FilterLogs(ctx, query)
		if err != nil {
			fmt.Printf("filter logs err height=%d, err=%v\n", currentHeight, err)
			currentHeight++
			time.Sleep(1 * time.Second)
			continue
		}

		var apiLogs []api.LogEvent
		for _, l := range logs {
			topics := make([]string, 0, len(l.Topics))
			for _, t := range l.Topics {
				topics = append(topics, t.Hex())
			}
			apiLogs = append(apiLogs, api.LogEvent{
				Address:     l.Address.Hex(),
				Topics:      topics,
				Data:        l.Data,
				BlockNumber: l.BlockNumber,
				TxHash:      l.TxHash.Hex(),
				LogIndex:    l.Index,
			})
		}

		if len(apiLogs) > 0 {
			evt := &api.ChainEvent{
				Height: currentHeight,
				Hash:   logs[0].BlockHash.Hex(),
				Logs:   apiLogs,
			}
			e.ch <- evt
		}

		currentHeight++
	}
}
