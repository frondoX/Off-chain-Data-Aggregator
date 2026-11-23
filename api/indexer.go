package api

import "context"

// ContractIndexer 合约索引插件
type ContractIndexer interface {
	Name() string
	Init(ctx context.Context, cfg *RuntimeConfig) error
	// Filter 声明自己关心哪些topic0，runtime用来做路由匹配
	Filter() []string
	// OnLog 收到一条匹配的日志，解码输出Record
	OnLog(ctx context.Context, log *LogEvent) ([]*Record, error)
}
