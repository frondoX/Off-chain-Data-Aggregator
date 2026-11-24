package retry

import (
	"context"
	"fmt"
	"time"
)

// Do 指数退避重试
// maxAttempt: 最大尝试次数（包含第一次）
// initialDelay: 初始等待时长
// fn: 要执行的业务函数
func Do(ctx context.Context, maxAttempt int, initialDelay time.Duration, fn func() error) error {
	if maxAttempt < 1 {
		return fmt.Errorf("maxAttempt must >=1")
	}
	var lastErr error
	delay := initialDelay
	for attempt := 0; attempt < maxAttempt; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		// ctx取消，直接退出，不再重试
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// 最后一次失败，不sleep
		if attempt == maxAttempt-1 {
			break
		}
		time.Sleep(delay)
		delay *= 2 // 指数退避
	}
	return fmt.Errorf("retry exhausted %d attempts, last err: %w", maxAttempt, lastErr)
}
