package pipeline

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// getCPUPercent получает процент использования CPU
func getCPUPercent() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	
	if len(percentages) > 0 {
		return percentages[0], nil
	}
	
	return 0, nil
}

// getMemoryInfo получает информацию о памяти
func getMemoryInfo() (map[string]interface{}, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":         memInfo.Total,
		"available":     memInfo.Available,
		"used":          memInfo.Used,
		"used_percent":  memInfo.UsedPercent,
		"free":          memInfo.Free,
		"cached":        memInfo.Cached,
		"buffers":       memInfo.Buffers,
	}, nil
}

// getNetworkStats получает статистику сети
func getNetworkStats() (map[string]interface{}, error) {
	netStats, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}

	if len(netStats) > 0 {
		stat := netStats[0]
		return map[string]interface{}{
			"bytes_sent":     stat.BytesSent,
			"bytes_recv":     stat.BytesRecv,
			"packets_sent":   stat.PacketsSent,
			"packets_recv":   stat.PacketsRecv,
			"errin":          stat.Errin,
			"errout":         stat.Errout,
			"dropin":         stat.Dropin,
			"dropout":        stat.Dropout,
		}, nil
	}

	return map[string]interface{}{}, nil
}

// getGoRuntimeStats получает статистику Go runtime
func getGoRuntimeStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines":        runtime.NumGoroutine(),
		"memory_alloc":      m.Alloc,
		"memory_total_alloc": m.TotalAlloc,
		"memory_sys":        m.Sys,
		"memory_lookups":    m.Lookups,
		"memory_mallocs":    m.Mallocs,
		"memory_frees":      m.Frees,
		"memory_heap_alloc": m.HeapAlloc,
		"memory_heap_sys":   m.HeapSys,
		"memory_heap_idle":  m.HeapIdle,
		"memory_heap_inuse": m.HeapInuse,
		"memory_heap_released": m.HeapReleased,
		"memory_heap_objects": m.HeapObjects,
		"memory_stack_inuse": m.StackInuse,
		"memory_stack_sys":   m.StackSys,
		"memory_gc_cycles":   m.NumGC,
		"memory_gc_pause_total": m.PauseTotalNs,
		"memory_gc_pause_avg":  m.PauseNs[(m.NumGC+255)%256],
	}
}
