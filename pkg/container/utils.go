package container

import (
	"fmt"
	"os"
)

func CheckRoot() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("This program must be run as root (use sudo or enter as root)")
	}
	return nil
}
