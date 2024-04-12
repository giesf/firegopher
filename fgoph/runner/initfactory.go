package runner

import (
	"log"
	"os"
	"os/exec"

	extfutils "firegopher.dev/extfutils"
)

/*
This generates the init file system.
The init file system only consists of two files:
- the run.toml containing all configuration for the workload
- the init binary that enforces this configuration
*/

func makeInitFS(initBinPath string, tomlFile string) (string, error) {
	// File to create
	ranId, err := makeId("init")

	fsFile := "/tmp/firegopher/" + ranId + ".ext4"
	createExt4Cmd := exec.Command("mke2fs", "-t", "ext4", fsFile, "4096")
	createExt4Cmd.Env = os.Environ()

	if err := createExt4Cmd.Run(); err != nil {
		log.Fatalf("Failed to create .ext4 file: %v", err)
		return fsFile, err
	}

	err = extfutils.MkdirInExt4(fsFile, "/firegopher")
	err = extfutils.CopyFileToExt4(fsFile, initBinPath, "/firegopher/guest-init")
	err = extfutils.CopyFileToExt4(fsFile, tomlFile, "/firegopher/run.toml")
	if err != nil {
		log.Fatalf("Failed to copy to .ext4 file: %v", err)
		return fsFile, err
	}

	log.Println("Successfully created .ext4 file with specified files.")
	return fsFile, nil
}
