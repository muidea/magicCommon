package os

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func GetContainerCPULimit() (float64, error) {
	// cgroup v2 路径：cpu.max 格式: "quota period"
	cpuMaxPath := "/sys/fs/cgroup/cpu.max"
	data, err := os.ReadFile(cpuMaxPath)
	if err == nil {
		parts := strings.Fields(string(data))
		if len(parts) == 2 {
			if parts[0] == "max" {
				return -1, nil
			}
			quota, err1 := strconv.ParseFloat(parts[0], 64)
			period, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 != nil || err2 != nil || period == 0 {
				return 0, fmt.Errorf("invalid cpu.max content")
			}
			return quota / period, nil
		}
	}

	// cgroup v1 路径
	cfsQuotaPath := "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	cfsPeriodPath := "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	quota, err1 := readFileToInt64(cfsQuotaPath)
	period, err2 := readFileToInt64(cfsPeriodPath)
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("failed to read cgroup v1 cpu files")
	}
	if quota == -1 {
		return -1, nil
	}
	if period == 0 {
		return 0, fmt.Errorf("cpu.cfs_period_us is zero")
	}
	return float64(quota) / float64(period), nil
}

// 获取系统CPU核数（非容器环境）
func GetSystemCPU() float64 {
	return float64(runtime.NumCPU())
}
