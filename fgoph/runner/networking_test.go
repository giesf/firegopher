package runner

import (
	"fmt"
	"testing"
)

func TestIpFinding(t *testing.T) {
	hip, gip, subnet, err := setupTapDeviceAndIpForwarding("tap-yeet", "enp0s25")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("IP", hip, gip, subnet, err)
}
