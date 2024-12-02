package ticker

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"testing"
	"time"
)

func TestComputeCPU(t *testing.T) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 获取CPU使用率
			cpuPercent, err := cpu.Percent(time.Second, false)
			if err != nil {
				fmt.Printf("Failed to get CPU usage: %s\n", err)
			} else {
				fmt.Printf("CPU Usage: %.2f%%\n", cpuPercent[0])
			}

			// 获取内存使用率
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				fmt.Printf("Failed to get memory usage: %s\n", err)
			} else {
				fmt.Printf("Memory Usage: %.2f\n", memInfo.UsedPercent)
			}

		}
	}
}
