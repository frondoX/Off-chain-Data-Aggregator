package worker

import (
	"context"
	"fmt"
	"time"

	"off-chain-data-aggregator/api"
)

type Runtime struct {
	cfg      *api.RuntimeConfig
	ds       api.DataSource
	indexers map[string]api.ContractIndexer
	sinks    []api.Sink
	// topic -> indexer列表，路由表
	routeTable map[string][]api.ContractIndexer
}

func NewRuntime(cfg *api.RuntimeConfig) *Runtime {
	return &Runtime{
		cfg:        cfg,
		indexers:   make(map[string]api.ContractIndexer),
		routeTable: make(map[string][]api.ContractIndexer),
	}
}

func (r *Runtime) Start(ctx context.Context) error {
	// 1. 加载数据源
	ds, ok := GetDataSource(r.cfg.Plugins.DataSource)
	if !ok {
		return fmt.Errorf("datasource %s not registered", r.cfg.Plugins.DataSource)
	}
	if err := ds.Init(ctx, r.cfg); err != nil {
		return err
	}
	r.ds = ds

	// 2. 加载indexers，构建topic路由表
	for _, idxName := range r.cfg.Plugins.Indexers {
		idx, ok := GetIndexer(idxName)
		if !ok {
			return fmt.Errorf("indexer %s not registered", idxName)
		}
		if err := idx.Init(ctx, nil); err != nil {
			return err
		}
		r.indexers[idxName] = idx
		for _, topic := range idx.Filter() {
			r.routeTable[topic] = append(r.routeTable[topic], idx)
		}
	}

	// 3. 加载sinks
	for _, sinkName := range r.cfg.Plugins.Sinks {
		sink, ok := GetSink(sinkName)
		if !ok {
			return fmt.Errorf("sink %s not registered", sinkName)
		}
		if err := sink.Init(ctx, r.cfg); err != nil {
			return err
		}
		r.sinks = append(r.sinks, sink)
	}

	// 4. 订阅流
	eventCh, err := r.ds.Subscribe(ctx)
	if err != nil {
		return err
	}

	// 5. 主循环
	go r.loop(ctx, eventCh)
	return nil
}

func (r *Runtime) loop(ctx context.Context, eventCh <-chan *api.ChainEvent) {
	maxBatch := r.cfg.MaxSinkBatchSize
	if maxBatch <= 0 {
		maxBatch = 2000
	}
	flushInterval := r.cfg.SinkFlushInterval
	if flushInterval <= 0 {
		flushInterval = time.Second
	}

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var allRecords []*api.Record

	for {
		select {
		case <-ctx.Done():
			// 退出时刷剩余数据
			if len(allRecords) > 0 {
				for _, s := range r.sinks {
					if err := s.Write(ctx, allRecords); err != nil {
						fmt.Printf("flush residual records err: %v, sink=%s\n", err, s.Name())
					}
				}
				allRecords = nil
			}
			return

		case <-ticker.C:
			// 定时刷盘：不够条数也落地
			if len(allRecords) > 0 {
				batch := allRecords
				for _, s := range r.sinks {
					if err := s.Write(ctx, batch); err != nil {
						fmt.Printf("sink flush interval err: %v, sink=%s\n", err, s.Name())
					}
				}
				allRecords = nil
			}

		case evt, ok := <-eventCh:
			if !ok {
				// 数据源通道关闭，刷残留
				if len(allRecords) > 0 {
					for _, s := range r.sinks {
						if err := s.Write(ctx, allRecords); err != nil {
							fmt.Printf("flush residual records err: %v, sink=%s\n", err, s.Name())
						}
					}
					allRecords = nil
				}
				return
			}

			for _, log := range evt.Logs {
				if len(log.Topics) == 0 {
					continue
				}
				topic0 := log.Topics[0]
				idxList, ok := r.routeTable[topic0]
				if !ok {
					continue
				}
				for _, idx := range idxList {
					recs, err := idx.OnLog(ctx, &log)
					if err != nil {
						fmt.Printf("indexer err: %v\n", err)
						continue
					}
					allRecords = append(allRecords, recs...)

					// 条数触发，满了立刻分批刷
					for len(allRecords) >= maxBatch {
						batch := allRecords[:maxBatch]
						for _, s := range r.sinks {
							if err := s.Write(ctx, batch); err != nil {
								fmt.Printf("sink write batch err: %v, sink=%s\n", err, s.Name())
							}
						}
						allRecords = allRecords[maxBatch:]
					}
				}
			}
		}
	}
}
