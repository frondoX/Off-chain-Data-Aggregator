package worker

import (
	"context"
	"fmt"

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
	fmt.Printf("DEBUG cfg raw: %#v\n", r.cfg)
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
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-eventCh:
			if !ok {
				return
			}
			var allRecords []*api.Record
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
						// demo：简单打日志，不做重试
						fmt.Printf("indexer err: %v\n", err)
						continue
					}
					allRecords = append(allRecords, recs...)
				}
			}
			if len(allRecords) == 0 {
				continue
			}
			// 写入所有sink
			for _, s := range r.sinks {
				_ = s.Write(ctx, allRecords)
			}
		}
	}
}
