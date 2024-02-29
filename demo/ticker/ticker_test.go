package ticker

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	// 间隔一秒钟的 ticker
	ticker := time.NewTicker(time.Second)
	// ticker 需要取消
	defer ticker.Stop()
	// 使用 ctx 控制住大循环
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for {
		// 使用 select 来监听 ticker
		select {
		case <-ctx.Done():
			t.Log("循环结束")
			return
		case <-ticker.C:
			t.Log("过了一秒")
		}
	}

}
