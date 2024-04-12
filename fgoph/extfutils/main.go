package extfutils

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"firegopher.dev/futils"
)

/*
	This is a go wrapper for the debugfs utility. Be sure to use debugfs >1.46 as older versions are buggy af and not compatible with this usecase
*/

func Debugfs(debugCmd string, ext4Path string) error {
	cmd := exec.Command("debugfs", "-w", ext4Path, "-R", debugCmd)
	cmd.Env = os.Environ()
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to run %s: %v", debugCmd, err)
		return err
	}
	cmd.Dir = path

	// sout, _ := cmd.StdoutPipe()
	// serr, _ := cmd.StderrPipe()
	// go io.Copy(os.Stdout, sout)
	// go io.Copy(os.Stdout, serr)

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run %s to .ext4 file: %v", debugCmd, err)
		return err
	}

	return nil
}

func CopyFileToExt4(ext4Path string, srcPath string, destPath string) error {

	fmt.Printf("Copying %s to %s in %s", srcPath, destPath, ext4Path)

	err := Debugfs("write "+srcPath+" "+destPath, ext4Path)

	return err

}
func MkdirInExt4(ext4Path string, destPath string) error {
	return Debugfs("mkdir "+destPath+"", ext4Path)
}

//@TODO unfuck this
func CopyDirToExt4(ext4Path string, srcPath string, destPath string) error {

	fmt.Printf("Copying %s to %s in %s", srcPath, destPath, ext4Path)

	MkdirInExt4(ext4Path, destPath)

	err := futils.WalkDirFast(srcPath, func(fileType, filePath string) {

		if fileType == "d" {
			MkdirInExt4(ext4Path, destPath+filePath)
		} else {
			CopyFileToExt4(ext4Path, srcPath+filePath, destPath+filePath)

		}
	})

	if err != nil {
		fmt.Println(err)
	}

	return err
}
