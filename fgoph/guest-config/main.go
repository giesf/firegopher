package guestconfig

import (
	"os"

	"firegopher.dev/futils"
	"github.com/pelletier/go-toml/v2"
)

/*
This is a library to parse and generate ron.toml configuration files
*/

type GuestConfig struct {
	Workload WorkloadConfig
	Security SecurityConfig
	Etc      EtcConfig
	Ip       IpConfig
}

type IpConfig struct {
	Ip      string
	Gateway string
	Mask    string
}

type EtcConfig struct {
	Hostname    string
	Hosts       []string
	Nameservers []string
}

type SecurityConfig struct {
	User  uint32
	Group uint32
}

type WorkloadConfig struct {
	Cmd  string
	Args []string
	Dir  string
}

func UnmarshalConfig(data []byte) (GuestConfig, error) {
	var conf GuestConfig
	err := toml.Unmarshal(data, &conf)
	return conf, err
}

func MarshalConfig(conf GuestConfig) ([]byte, error) {
	return toml.Marshal(conf)
}

func ReadConfigFromFile(filePath string) (GuestConfig, error) {
	var conf GuestConfig
	rawConfig, err := os.ReadFile(filePath)
	if err != nil {
		return conf, err
	}
	conf, err = UnmarshalConfig(rawConfig)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func SaveConfigToFile(conf GuestConfig, filePath string) error {
	rawConfig, err := MarshalConfig(conf)
	if err != nil {
		return err
	}
	confStr := string(rawConfig)

	err = futils.CreateFileWithContent(filePath, confStr)
	if err != nil {
		return err
	}

	return nil
}
