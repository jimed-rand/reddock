package utils

import (
	"os"
	"runtime"
	"strings"
)

type CPUVendor string

const (
	VendorIntel CPUVendor = "intel"
	VendorAMD   CPUVendor = "amd"
	VendorOther CPUVendor = "other"
)

func GetCPUVendor() CPUVendor {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return VendorOther
	}

	content := string(data)
	if strings.Contains(content, "GenuineIntel") {
		return VendorIntel
	} else if strings.Contains(content, "AuthenticAMD") {
		return VendorAMD
	}

	return VendorOther
}
func (v CPUVendor) String() string {
	return string(v)
}

func IsARM() bool {
	arch := runtime.GOARCH
	return arch == "arm" || arch == "arm64"
}
