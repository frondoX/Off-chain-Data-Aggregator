package api

import "time"

// ChainConfig 单链配置
type ChainConfig struct {
	ChainID     uint64 `yaml:"chain_id"`
	StartHeight uint64 `yaml:"start_height"`
	RPCEndpoint string `yaml:"rpc_endpoint"`
}

// PluginConfig 插件通用配置，每个插件可以拿到自己的子配置
type PluginConfig map[string]any

// RuntimeConfig 顶层配置
type RuntimeConfig struct {
	Chain             ChainConfig       `yaml:"chain"`
	Plugins           PluginsSetting    `yaml:"plugins"`
	Clickhouse        *ClickhouseConfig `yaml:"clickhouse"`
	MaxSinkBatchSize  int               `yaml:"max_sink_batch_size"` // 单批最大记录数，例如 2000
	SinkFlushInterval time.Duration     `yaml:"sink_flush_interval"`
}

type PluginsSetting struct {
	DataSource string   `yaml:"datasource"`
	Indexers   []string `yaml:"indexers"`
	Sinks      []string `yaml:"sinks"`
}

type ClickhouseConfig struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}
