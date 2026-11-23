package api

import "context"

// Sink 输出插件
type Sink interface {
	Name() string
	Init(ctx context.Context, cfg *RuntimeConfig) error
	// Write 批量写入记录
	Write(ctx context.Context, records []*Record) error
}
