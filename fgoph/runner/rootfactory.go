package runner

import (
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"firegopher.dev/extfutils"
	"firegopher.dev/futils"
)

func makeRootFs(runChecksum string, appZip string, base string, assetBasePath string) (string, error) {

	ranId, _ := makeId("root")
	fsFile := "/tmp/firegopher/root_" + runChecksum + ".ext4"

	if _, err := os.Stat(fsFile); err == nil {
		// The file exists @TODO deal with that?
		return fsFile, nil
	}

	futils.CopyFile(assetBasePath+"/"+base+".ext4", fsFile)

	if appZip == "NONE" {
		return fsFile, nil
	}

	appPrepDir := "/tmp/firegopher/app_" + runChecksum + ranId
	futils.Mkdir(appPrepDir, 0o755)

	startUnzip := time.Now()
	unzipCmd := exec.Command("unzip", appZip, "-d", appPrepDir)
	unzipCmd.Env = os.Environ()
	sout, _ := unzipCmd.StdoutPipe()
	serr, _ := unzipCmd.StderrPipe()
	go io.Copy(os.Stdout, sout)
	go io.Copy(os.Stdout, serr)

	if err := unzipCmd.Run(); err != nil {
		log.Fatalf("Failed to create root.ext4 file: %v", err)
		return fsFile, err
	}
	endUnzip := time.Now()
	printDifference("UNZIP", startUnzip, endUnzip)

	err := extfutils.CopyDirToExt4(fsFile, appPrepDir, "/app")

	return fsFile, err
}
