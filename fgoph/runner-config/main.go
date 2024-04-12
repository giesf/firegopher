package runnerconfig

type RunnerConfig struct {
	CheckSum   string
	Quiet      bool
	Workload   WorkloadConfig
	Os         OSConfig
	Networking NetworkingConfig
}

type OSConfig struct {
	BaseImage string
}

type WorkloadConfig struct {
	AppZip     string
	DataVolume string
	Cmd        string
	Args       []string
	Dir        string
}

type NetworkingConfig struct {
	TapDeviceName string
}
