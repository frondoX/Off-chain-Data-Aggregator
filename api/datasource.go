package api

import "context"

// DataSource 数据源插件接口
type DataSource interface {
	// Name 插件标识，和配置里名字对应
	Name() string
	// Init 初始化，传入该插件的配置
	Init(ctx context.Context, cfg *RuntimeConfig) error
	// Subscribe 阻塞返回事件流
	Subscribe(ctx context.Context) (<-chan *ChainEvent, error)
}
