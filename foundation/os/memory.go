package os

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 判断是否在容器内运行
func IsRunningInContainer() (bool, error) {
	cgHandle, cgErr := os.Open("/proc/1/cgroup")
	if cgErr != nil {
		return false, cgErr
	}
	defer cgHandle.Close()

	scanner := bufio.NewScanner(cgHandle)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "docker") ||
			strings.Contains(line, "kubepods") ||
			strings.Contains(line, "containerd") ||
			strings.Contains(line, "cri-containerd") {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func readFileToInt64(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	if s == "max" {
		return -1, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

// 读取内存限制，返回字节数
func GetContainerMemoryLimit() (int64, error) {
	// cgroup v2 路径
	memMaxPath := "/sys/fs/cgroup/memory.max"
	if limit, err := readFileToInt64(memMaxPath); err == nil {
		return limit, nil
	}
	// cgroup v1 路径
	memLimitPath := "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	return readFileToInt64(memLimitPath)
}

// 非容器环境读取系统总内存，单位字节
func GetSystemMemory() (int64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0, fmt.Errorf("invalid MemTotal line: %s", line)
			}
			// MemTotal 单位是 KB，转换成字节
			kbValue, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return kbValue * 1024, nil
		}
	}
	return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
}

/*
func main() {
	inContainer, err := IsRunningInContainer()
	if err != nil {
		fmt.Printf("Failed to check container environment: %v\n", err)
		return
	}

	var memBytes int64
	if inContainer {
		memBytes, err = GetContainerMemoryLimit()
		if err != nil {
			fmt.Printf("Error reading container memory limit: %v\n", err)
			return
		}
		if memBytes == -1 {
			fmt.Println("Memory limit: unlimited")
		} else {
			safeMemMB := float64(memBytes) * 0.8 / (1024 * 1024)
			fmt.Printf("Container memory limit with 0.8 factor: %.2f MB\n", safeMemMB)
		}
	} else {
		memBytes, err = GetSystemMemory()
		if err != nil {
			fmt.Printf("Error reading system memory: %v\n", err)
			return
		}
		safeMemMB := float64(memBytes) * 0.8 / (1024 * 1024)
		fmt.Printf("System memory with 0.8 factor: %.2f MB\n", safeMemMB)
	}
}
*/
