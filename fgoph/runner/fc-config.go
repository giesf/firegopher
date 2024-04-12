package runner

/*
This is a factory to generate Firecracker JSON configuration files
Check https://github.com/firecracker-microvm/firecracker/blob/main/tests/framework/vm_config.json
Or https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md#configuring-the-microvm-without-sending-api-requests
*/

type Config struct {
	BootSource        BootSource         `json:"boot-source"`
	Drives            []Drive            `json:"drives"`
	NetworkInterfaces []NetworkInterface `json:"network-interfaces"`
	MachineConfig     MachineConfig      `json:"machine-config"`
}

type BootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args"`
}

type Drive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

type NetworkInterface struct {
	IfaceID     string `json:"iface_id"`
	HostDevName string `json:"host_dev_name"`
	GuestMAC    string `json:"guest_mac"`
}

type MachineConfig struct {
	VcpuCount  int `json:"vcpu_count"`
	MemSizeMiB int `json:"mem_size_mib"`
}

func makeFirecrackerConfig(hostDevName string, guestMAC string, includeDataVolume bool) Config {

	drives := []Drive{
		{
			DriveID:      "initfs",
			PathOnHost:   "/init",
			IsRootDevice: true,
			IsReadOnly:   false,
		},
		{
			DriveID:      "rootfs",
			PathOnHost:   "/rootfs.ext4",
			IsRootDevice: false,
			IsReadOnly:   false,
		},
	}

	if includeDataVolume {
		drives = append(drives, Drive{
			DriveID:      "uservolume",
			PathOnHost:   "/dataVolume.ext4",
			IsRootDevice: false,
			IsReadOnly:   false,
		})
	}

	return Config{
		BootSource: BootSource{
			KernelImagePath: "./kernel",
			BootArgs:        "ro console=ttyS0 noapic reboot=k panic=1 pci=off nomodules random.trust_cpu=on init=/firegopher/guest-init",
		},
		Drives: drives,
		NetworkInterfaces: []NetworkInterface{
			{
				IfaceID:     "eth0",
				HostDevName: hostDevName,
				GuestMAC:    guestMAC,
			},
		},
		MachineConfig: MachineConfig{
			VcpuCount:  1,
			MemSizeMiB: 128,
		},
	}
}
