package worker

import (
	"off-chain-data-aggregator/api"
)

// DataSourceFactory 数据源工厂
type DataSourceFactory func() api.DataSource

// IndexerFactory 索引插件工厂
type IndexerFactory func() api.ContractIndexer

// SinkFactory sink工厂
type SinkFactory func() api.Sink

var (
	dsRegistry   = make(map[string]DataSourceFactory)
	idxRegistry  = make(map[string]IndexerFactory)
	sinkRegistry = make(map[string]SinkFactory)
)

func RegisterDataSource(name string, f DataSourceFactory) {
	dsRegistry[name] = f
}
func RegisterIndexer(name string, f IndexerFactory) {
	idxRegistry[name] = f
}
func RegisterSink(name string, f SinkFactory) {
	sinkRegistry[name] = f
}

// Getters runtime启动时根据配置名字拿
func GetDataSource(name string) (api.DataSource, bool) {
	f, ok := dsRegistry[name]
	if !ok {
		return nil, false
	}
	return f(), true
}
func GetIndexer(name string) (api.ContractIndexer, bool) {
	f, ok := idxRegistry[name]
	if !ok {
		return nil, false
	}
	return f(), true
}
func GetSink(name string) (api.Sink, bool) {
	f, ok := sinkRegistry[name]
	if !ok {
		return nil, false
	}
	return f(), true
}
