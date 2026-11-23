package mock_ds

import (
	"context"
	"off-chain-data-aggregator/api"
	"off-chain-data-aggregator/worker"
	"time"
)

func init() {
	worker.RegisterDataSource("mock", func() api.DataSource {
		return &MockDS{}
	})
}

type MockDS struct {
	ch chan *api.ChainEvent
}

func (m *MockDS) Name() string {
	return "mock"
}

func (m *MockDS) Init(ctx context.Context, cfg *api.RuntimeConfig) error {
	m.ch = make(chan *api.ChainEvent, 8)
	go m.produce(ctx)
	return nil
}

func (m *MockDS) Subscribe(ctx context.Context) (<-chan *api.ChainEvent, error) {
	return m.ch, nil
}

func (m *MockDS) produce(ctx context.Context) {
	height := uint64(20000000)
	swapTopic := "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(m.ch)
			return
		case <-ticker.C:
			evt := &api.ChainEvent{
				Height: height,
				Hash:   "0xmockblockhash",
				Logs: []api.LogEvent{
					{
						Address:     "0xmockpair",
						Topics:      []string{swapTopic},
						Data:        []byte("0xmockdata"),
						BlockNumber: height,
						TxHash:      "0xmocktx",
						LogIndex:    0,
					},
				},
			}
			m.ch <- evt
			height++
		}
	}
}
