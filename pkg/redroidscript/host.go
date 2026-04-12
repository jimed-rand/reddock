package redroidscript

import (
	"fmt"
	"runtime"
)

// HostTuple matches ayasa520/redroid-script tools.helper.host(): (archKey, bits).
func HostTuple() (arch string, bits int, err error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", 64, nil
	case "386":
		return "x86", 32, nil
	case "arm64":
		return "arm64", 64, nil
	case "arm":
		return "arm", 32, nil
	default:
		return "", 0, fmt.Errorf("unsupported GOARCH %q for redroid add-on builds", runtime.GOARCH)
	}
}
