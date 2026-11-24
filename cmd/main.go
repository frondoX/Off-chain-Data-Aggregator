package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"off-chain-data-aggregator/api"
	"off-chain-data-aggregator/worker"

	"gopkg.in/yaml.v3"

	// 重要：导入插件包，触发init()注册
	_ "off-chain-data-aggregator/plugins/datasource/evm_rpc"
	_ "off-chain-data-aggregator/plugins/datasource/mock_ds"
	_ "off-chain-data-aggregator/plugins/indexer/uniswap_v2"
	_ "off-chain-data-aggregator/plugins/sink/clickhouse"
	_ "off-chain-data-aggregator/plugins/sink/stdout"
)

func main() {
	cfgPath := flag.String("config", "configs/demo.yaml", "config file path")
	flag.Parse()

	data, err := os.ReadFile(*cfgPath)
	if err != nil {
		fmt.Printf("read config err: %v\n", err)
		os.Exit(1)
	}
	var cfg api.RuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("parse yaml err: %v\n", err)
		os.Exit(1)
	}

	// 创建顶层上下文，确保关闭时候，所有goroutine都能收到通知
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nreceive shutdown signal")
		cancel()
	}()

	// 创建runtime并启动
	rt := worker.NewRuntime(&cfg)
	if err := rt.Start(ctx); err != nil {
		fmt.Printf("start runtime err: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	fmt.Println("shutdown complete")
}
