package runner

import (
	"net"
	"strings"

	guestconfig "firegopher.dev/guest-config"
	runnerconfig "firegopher.dev/runner-config"
)

func makeGuestConfig(rc runnerconfig.RunnerConfig, gatewayIP net.IP, guestIP net.IP, subnet *net.IPNet) (guestConfigPath string, err error) {
	baseConfigPath := "/tmp/firegopher"
	configFileName := "rc_" + rc.CheckSum + ".toml"

	shortMask := strings.Split(subnet.String(), "/")[1]

	guestConfig := guestconfig.GuestConfig{
		Workload: guestconfig.WorkloadConfig{
			Cmd:  rc.Workload.Cmd,
			Args: rc.Workload.Args,
			Dir:  rc.Workload.Dir,
		},
		//@TODO make this make sense
		Security: guestconfig.SecurityConfig{
			User:  123,
			Group: 123,
		},
		Etc: guestconfig.EtcConfig{
			Hostname:    "miau",
			Hosts:       []string{},
			Nameservers: []string{"8.8.8.8"},
		},
		Ip: guestconfig.IpConfig{
			Ip:      guestIP.String(),
			Gateway: gatewayIP.String(),
			Mask:    shortMask,
		},
	}

	guestConfigPath = baseConfigPath + "/" + configFileName

	guestconfig.SaveConfigToFile(guestConfig, guestConfigPath)

	return
}
