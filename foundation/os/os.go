package os

import (
	"os"
	"runtime"
)

func GetOsName() string {
	return runtime.GOOS
}

func GetDefaultShell() string {
	return os.Getenv("SHELL")
}

func GetCurrentWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// 获取当前用户家目录
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func GetMemoryUseable() (int64, error) {
	if IsRunningInContainer() {
		return GetContainerMemoryLimit()
	}

	return GetSystemMemory()
}

func GetCPUUseable() (float64, error) {
	if IsRunningInContainer() {
		return GetContainerCPULimit()
	}

	return GetSystemCPU(), nil
}
