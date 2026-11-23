package api

// ChainEvent 统一事件载体，demo只用到LogEvent
type ChainEvent struct {
	Height uint64
	Hash   string
	Logs   []LogEvent
}

// LogEvent 合约日志
type LogEvent struct {
	Address     string
	Topics      []string
	Data        []byte
	BlockNumber uint64
	TxHash      string
	LogIndex    uint
}

// Record indexer输出的结构化业务记录，交给sink写入
type Record struct {
	Table string                 `json:"table"`
	Data  map[string]interface{} `json:"data"`
}
