CREATE TABLE IF NOT EXISTS uni_v2_swap (
    block_number UInt64,
    tx_hash String,
    log_index UInt32,
    pair String,
    sender String,
    to String,
    amount0In String,
    amount1In String,
    amount0Out String,
    amount1Out String,
    raw_data String,
    create_time DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (block_number, tx_hash, log_index);