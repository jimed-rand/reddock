package container

import (
	"fmt"
	"os"
)

func CheckRoot() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("this command must be run as root (use sudo)")
	}
	return nil
}
