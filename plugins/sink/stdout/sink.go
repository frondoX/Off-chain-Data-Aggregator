package stdout

import (
	"context"
	"encoding/json"
	"fmt"
	"off-chain-data-aggregator/api"
	"off-chain-data-aggregator/worker"
)

func init() {
	worker.RegisterSink("stdout", func() api.Sink {
		return &Sink{}
	})
}

type Sink struct{}

func (s *Sink) Name() string {
	return "stdout"
}

func (s *Sink) Init(ctx context.Context, cfg *api.RuntimeConfig) error {
	return nil
}

func (s *Sink) Write(ctx context.Context, records []*api.Record) error {
	b, _ := json.MarshalIndent(records, "", "  ")
	fmt.Println("\n===== Sink Output =====")
	fmt.Println(string(b))
	return nil
}
