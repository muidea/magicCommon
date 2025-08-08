package os

import (
	"bufio"
	"os"
	"strings"
)

func IsRunningInContainer() bool {
	// 1. 检查特定文件
	if fileExists("/.dockerenv") || fileExists("/run/.containerenv") {
		return true
	}

	// 2. 检查 /proc/1/cgroup
	if inContainerByCgroup() {
		return true
	}

	// 3. 检查 /proc/1/sched 是否不是 systemd/init
	if inContainerBySched() {
		return true
	}

	// 4. 检查 /proc/self/mountinfo 是否包含 overlay
	if inContainerByMountInfo() {
		return true
	}

	return false
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func inContainerByCgroup() bool {
	file, err := os.Open("/proc/1/cgroup")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "docker") ||
			strings.Contains(line, "containerd") ||
			strings.Contains(line, "podman") ||
			strings.Contains(line, "kubepods") {
			return true
		}
	}
	return false
}

func inContainerBySched() bool {
	data, err := os.ReadFile("/proc/1/sched")
	if err != nil {
		return false
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	// 宿主机 PID 1 一般是 systemd/init
	if !strings.Contains(line, "systemd") && !strings.Contains(line, "init") {
		return true
	}
	return false
}

func inContainerByMountInfo() bool {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "overlay") {
			return true
		}
	}
	return false
}
