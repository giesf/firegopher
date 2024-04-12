package assetmanager

import (
	"io"
	"log"
	"os"
	"os/exec"

	"firegopher.dev/futils"
)

var arch = "x86_64"
var firecrackerVersion = "v1.6.0"
var kernelDownloadUrl = "https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.5/" + arch + "/vmlinux-5.10.186"
var bucketUrl = "https://firegopher-assets.s3.fr-par.scw.cloud"

var tmpDir = "/tmp/firegopher"
var srvDir = "/srv/firegopher"

func DownloadAssets() {
	initBinName := "guest-init"
	rootFsName := "ubuntu.ext4"
	initBinDownloadUrl := bucketUrl + "/" + initBinName
	rootFsDownloadUrl := bucketUrl + "/" + rootFsName

	futils.Mkdir(srvDir, 0o755)
	futils.Mkdir(srvDir+"/bin", 0o755)

	runCmd(srvDir+"/bin", "curl", "-o", initBinName, initBinDownloadUrl)
	runCmd(srvDir+"/bin", "chmod", "+x", initBinName)
	runCmd(srvDir, "curl", "-o", rootFsName, rootFsDownloadUrl)
	runCmd(srvDir, "curl", "-o", "kernel", kernelDownloadUrl)
}

func InstallFirecracker() {
	version := firecrackerVersion
	suffix := "-" + version + "-" + arch

	releaseUrl := "https://github.com/firecracker-microvm/firecracker/releases"
	fileName := "firecracker" + suffix + ".tgz"
	downloadUrl := releaseUrl + "/download/" + version + "/" + fileName
	installDir := srvDir
	binDir := "/usr/bin"
	extractedDir := "release" + suffix
	jailerBin := "jailer" + suffix
	firecrackerBin := "firecracker" + suffix

	downloadDest := installDir + "/" + fileName

	log.Println(downloadUrl)

	log.Println(downloadDest)

	futils.Mkdir(installDir, 0o755)

	log.Println("Downloading...")
	runCmd("", "curl", "-o", downloadDest, "-L", downloadUrl)
	log.Println("Extracting...")
	log.Println("Changing directory...", installDir)

	runCmd(installDir, "tar", "-xzf", fileName)

	log.Println("Linking...")

	runCmd(installDir, "ln", "-sfn", installDir+"/"+extractedDir+"/"+firecrackerBin, binDir+"/firecracker")
	runCmd(installDir, "ln", "-sfn", installDir+"/"+extractedDir+"/"+jailerBin, binDir+"/jailer")

	runCmd("", "firecracker", "--version")
	runCmd("", "jailer", "--version")

}

func runCmd(cwd string, bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Env = os.Environ()
	var err error
	path := cwd
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to run %s: %v", bin, err)
			return err
		}
	}
	cmd.Dir = path

	sout, _ := cmd.StdoutPipe()
	serr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, sout)
	go io.Copy(os.Stdout, serr)

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run %s: %v", bin, err)
		return err
	}

	return nil
}
