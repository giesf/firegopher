package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"firegopher.dev/futils"
	runnerconfig "firegopher.dev/runner-config"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

/*
This is a basic, jailed firecracker supervisor.
- It prepares a root file system, a kernel and an init system for the firecracker process
- It configures a network device for the firecracker process
- It copies the prepared assets to a jail root directory
- It starts firecracker as a jailed process


*/
//@TODO Allow for IP to be specified
//@TODO automatically choose internet-facing network device if not specified
//@TODO Think about the device clean up and other shutdown tasks
//@TODO Add way to clean up old jails
//@TODO Think about error handling a bit more
//@TODO clean up the types of random ids used.
//@TODO Remove hardcoded firecracker path

func Runner(runConfig runnerconfig.RunnerConfig) {

	fmt.Println("Checksum", runConfig.CheckSum)

	jailId, _ := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz123456789-", 64)

	os.Mkdir("/tmp/firegopher", 0o755)

	assetBasePath := "/srv/firegopher"

	kernalImagePath := assetBasePath + "/kernel"
	initBinaryPath := assetBasePath + "/bin/guest-init"

	fmt.Println("Creating root FS...")
	rootFSPath, err := makeRootFs(runConfig.CheckSum, runConfig.Workload.AppZip, runConfig.Os.BaseImage, assetBasePath)
	if err != nil {
		fmt.Println("Error creating root fs", err)
		return
	}

	// Networking

	guestMAC, err := generateMAC("02:FC")
	if err != nil {
		fmt.Println("Error generating MAC", err)
		return
	}

	ifName := "ens2" // "enp0s25"
	gatewayIP, guestIP, subnet, err := setupTapDeviceAndIpForwarding(runConfig.Networking.TapDeviceName, ifName)
	if err != nil {
		fmt.Println(err)
		return
	}

	guestConfigFilePath, err := makeGuestConfig(runConfig, gatewayIP, guestIP, subnet)
	if err != nil {
		fmt.Println("Error creating guest config file", err)
		return
	}

	initFsPath, err := makeInitFS(initBinaryPath, guestConfigFilePath)
	if err != nil {
		fmt.Println("Error creating init fs", err)
		return
	}

	jailerRootBase := "/srv/jailer"
	jailerRoot := jailerRootBase + "/firecracker-v1.6.0-x86_64/"
	baseDir := jailerRoot + jailId
	tmpDir := baseDir + "/root"

	os.Mkdir(jailerRootBase, 0o755)
	os.Mkdir(jailerRoot, 0o755)

	os.Mkdir(baseDir, 0o755)
	os.Mkdir(tmpDir, 0o755)

	jailedKernelPath := tmpDir + "/kernel"
	jailedRootFSPath := tmpDir + "/rootfs.ext4"
	jailedInitFSPath := tmpDir + "/init"

	futils.CopyFile(kernalImagePath, jailedKernelPath)
	futils.CopyFile(rootFSPath, jailedRootFSPath)
	futils.CopyFile(initFsPath, jailedInitFSPath)

	os.Chown(jailedKernelPath, 123, 100)
	os.Chown(jailedRootFSPath, 123, 100)
	os.Chown(jailedInitFSPath, 123, 100)

	configFilePath := tmpDir + "/vm-config.json"
	dataVolume := runConfig.Workload.DataVolume
	if dataVolume != "NONE" {
		jailedDataVolume := tmpDir + "/dataVolume.ext4"

		linkErr := os.Link(dataVolume, jailedDataVolume)
		if linkErr != nil {
			fmt.Println(err)
			return
		}
		os.Chown(jailedDataVolume, 123, 100)
	}

	cfg := makeFirecrackerConfig(runConfig.Networking.TapDeviceName, guestMAC, dataVolume != "NONE")
	cfgString, err := json.Marshal(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	futils.CreateFileWithContent(configFilePath, string(cfgString))
	os.Chown(configFilePath, 123, 100)

	firecrackerBinary := "/usr/bin/firecracker"
	jailerBinary := "jailer"

	//@TODO select these more carefully?
	uid := "123"
	gid := "100"

	fmt.Println("Starting Firecracker...")
	fmt.Println("\n\n############# START OF FIRECRACKER JAILER OUTPUT #############")
	execute(jailerBinary, []string{"--exec-file", firecrackerBinary, "--id", jailId, "--chroot-base-dir", jailerRootBase, "--uid", uid, "--gid", gid, "--", "--config-file", "/vm-config.json"}, ".", 0, 0, runConfig.Quiet)

}

func printDifference(label string, t1 time.Time, t2 time.Time) {
	fmt.Printf("[TIME DIFF] %s: %s\n", label, t2.Sub(t1).String())
}
