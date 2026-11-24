package clickhouse

import (
	"context"
	"fmt"
	"time"

	"off-chain-data-aggregator/api"
	"off-chain-data-aggregator/pkg/retry"
	"off-chain-data-aggregator/worker"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func init() {
	worker.RegisterSink("clickhouse", func() api.Sink {
		return &ClickHouseSink{}
	})
}

type ClickHouseSink struct {
	conn driver.Conn
}

func (s *ClickHouseSink) Name() string {
	return "clickhouse"
}

func (s *ClickHouseSink) Init(ctx context.Context, cfg *api.RuntimeConfig) error {

	chCfg := cfg.Clickhouse
	if chCfg == nil {
		return fmt.Errorf("clickhouse config not found in yaml")
	}
	// Init阶段：连接+ping，重试
	err := retry.Do(ctx, 3, 200*time.Millisecond, func() error {
		conn, err := clickhouse.Open(&clickhouse.Options{
			Addr: []string{chCfg.Addr},
			Auth: clickhouse.Auth{
				Database: chCfg.DB,
				Username: chCfg.User,
				Password: chCfg.Password,
			},
		})
		if err != nil {
			return err
		}
		if err := conn.Ping(ctx); err != nil {
			return err
		}
		s.conn = conn
		return nil
	})
	if err != nil {
		return fmt.Errorf("ch init retry failed: %w", err)
	}

	fmt.Printf("[CH Sink] init success, addr=%s, db=%s\n", chCfg.Addr, chCfg.DB)
	return nil
}

func (s *ClickHouseSink) Write(ctx context.Context, recs []*api.Record) error {
	if len(recs) == 0 {
		fmt.Println("[CH Sink] recs empty, skip")
		return nil
	}
	fmt.Printf("[CH Sink] receive %d records\n", len(recs))

	// Write写入重试
	err := retry.Do(ctx, 3, 200*time.Millisecond, func() error {
		batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO uni_v2_swap (block_number,tx_hash,log_index,pair,sender,to,amount0In,amount1In,amount0Out,amount1Out,raw_data)")
		if err != nil {
			fmt.Printf("[CH Sink] prepare batch err: %v\n", err)
			return err
		}
		for _, r := range recs {
			data := r.Data
			err = batch.Append(
				data["block_number"],
				data["tx_hash"],
				data["log_index"],
				data["pair"],
				data["sender"],
				data["to"],
				data["amount0In"],
				data["amount1In"],
				data["amount0Out"],
				data["amount1Out"],
				data["raw_data"],
			)
			if err != nil {
				fmt.Printf("[CH Sink] append err: %v, data: %+v\n", err, data)
				_ = batch.Abort()
				return err
			}
		}
		err = batch.Send()
		if err != nil {
			fmt.Printf("[CH Sink] batch send err: %v\n", err)
			_ = batch.Abort()
			return err
		}
		fmt.Println("[CH Sink] batch send success")
		return nil
	})
	return err
}
