# Off-chain-Data-Aggregator
这是一个区块链链下数据索引器


indexer-platform/
├── go.mod
├── go.sum
├── Makefile                  # 编译、生成proto、lint、e2e、本地单节点启动
├── configs/
│   ├── example.evm.yaml      # 示例配置：ETH主网，firehose+archive_rpc，uniswap v2/v3
│   └── schema.yaml            # 配置json schema（可选，校验）
├── api/                      # 【最核心：所有抽象接口、公共模型、proto】
│   ├── event.go              # ChainEvent、LogEvent、TraceEvent、StateDiffEvent、UndoMark
│   ├── datasource.go         # DataSource 接口定义
│   ├── indexer.go            # ContractIndexer 接口定义
│   ├── sink.go               # Sink 接口定义
│   ├── checkpoint.go         # Checkpoint 结构体、FencingToken 定义
│   ├── config.go             # 配置结构体（ChainConfig、PluginEnable、SinkConfig）
│   └── proto/                # 后续跨进程通信proto（控制面<->worker）
│       └── control.proto
├── pkg/                      # 公共工具库，无业务、无强依赖
│   ├── logger/               # zap封装，统一日志、trace注入
│   ├── metrics/              # prometheus指标：块延迟、reorg次数、插件error、sink写入耗时
│   ├── otel/                 # trace链路
│   ├── backoff/              # 重试退避
│   ├── queue/                # 内部无锁分片队列、背压队列
│   ├── hashutil/
│   └── codec/                # ABI辅助工具、hex、rlp等通用工具（不写具体合约解码）
├── controlplane/             # 【第一层：控制平面】独立模块，可以单独编译成二进制
│   ├── app.go                # control-plane主入口
│   ├── config_manager.go     # 配置加载、热加载、版本管理
│   ├── scheduler.go          # 分片调度、分配策略、扩缩容感知
│   ├── etcd_store.go         # etcd元数据、分片锁、fencing token发放
│   ├── watcher.go            # watch etcd分片变更、配置变更
│   ├── server.go             # http/grpc 管控接口：健康、配置推送、分片查询
│   └── notify.go             # 通知worker分片变更、token更新
├── worker/                   # 【第二层：核心编排运行时，Worker内核】
│   ├── app.go                # worker进程主入口
│   ├── runtime.go            # Runtime 主结构体，编排核心
│   ├── stream_mgr.go         # 数据源流管理器，多数据源主备切换
│   ├── reorg_engine.go       # 分叉检测、确认窗口、undo事件分发【重点】
│   ├── shard_router.go       # 分片路由，事件分发到indexer，分片内保序
│   ├── checkpoint_store.go   # checkpoint读写 + fencing token校验逻辑
│   ├── limiter.go            # 限流、背压控制
│   └── registry.go           # 插件注册中心（全局，注册所有datasource/indexer/sink）
├── plugins/                  # 【第三层：所有插件实现，只依赖 api 包】
│   ├── datasource/           # 数据源插件实现
│   │   ├── evm_firehose/
│   │   │   └── firehose.go
│   │   ├── evm_archive_rpc/
│   │   │   └── archive_rpc.go
│   │   └── solana_geyser/    # 后续接入，现在先占位
│   │       └── geyser.go
│   ├── indexer/              # 合约索引插件（你说的uniswapV2/V3在这里）
│   │   ├── uniswap_v2/
│   │   │   ├── abi.go
│   │   │   ├── indexer.go    # 实现 api.ContractIndexer
│   │   │   └── model.go      # 业务结构体：Swap, Mint, Burn
│   │   ├── uniswap_v3/
│   │   │   ├── abi.go
│   │   │   ├── indexer.go
│   │   │   └── model.go
│   │   ├── erc20/
│   │   └── erc721/
│   └── sink/                 # Sink输出插件，同样只依赖api
│       ├── kafka/
│       ├── clickhouse/
│       ├── postgres/
│       ├── redis/
│       └── s3/
├── cmd/                      # 二进制入口，一个子命令对应一个进程
│   ├── root.go
│   ├── controlplane.go       # ./indexer-platform control-plane
│   ├── worker.go             # ./indexer-platform worker
│   └── dev.go                # 本地调试：单进程模式（controlplane + worker 合并跑）
├── internal/                 # 内部私有包，只给本项目用，不对外暴露
│   ├── ringbuffer/
│   └── semaphore/
├── tests/
│   ├── e2e/                  # 端到端测试：本地私链，跑UniV2索引
│   └── mock/                 # mock datasource、mock indexer、mock sink
└── docs/
    ├── architecture.md
    ├── plugin-dev.md         # 插件开发文档：如何写一个新indexer
    └── fencing-token-flow.md



最简版本
indexer-platform/
├── go.mod
├── go.sum
├── Makefile
├── configs/
│   └── demo.yaml
├── api/                     # 契约不动，保留最小接口
│   ├── event.go
│   ├── datasource.go
│   ├── indexer.go
│   ├── sink.go
│   └── config.go
├── pkg/
│   ├── logger/
│   └── metrics/             # 极简metrics，只统计块数、解码失败数
├── worker/                  # 简化runtime，没有shard、没有reorg_engine
│   ├── app.go
│   ├── runtime.go           # 简单流调度、插件路由
│   └── registry.go          # 插件注册中心
├── plugins/
│   ├── datasource/
│   │   └── evm_archive_rpc/
│   ├── indexer/
│   │   └── uniswap_v2/
│   └── sink/
│       └── clickhouse/
├── cmd/
│   └── root.go              # 只有一个命令：demo 启动
└── tests/
    └── mock
