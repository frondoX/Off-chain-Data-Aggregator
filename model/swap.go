package model

type Swap struct {
	ChainID     uint64 `ch:"chain_id"`
	BlockNumber uint64 `ch:"block_number"`
	TxHash      string `ch:"tx_hash"`
	LogIndex    uint32 `ch:"log_index"`
	Pair        string `ch:"pair"`
	Sender      string `ch:"sender"`
	To          string `ch:"to"`
	Amount0In   string `ch:"amount0_in"`
	Amount1In   string `ch:"amount1_in"`
	Amount0Out  string `ch:"amount0_out"`
	Amount1Out  string `ch:"amount1_out"`
	RawData     string `ch:"raw_data"`
}
