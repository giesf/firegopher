package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

func execute(execPath string, args []string, dir string, uid uint32, gid uint32, quiet bool) error {

	cmd := exec.Command(execPath, args...)

	sout, _ := cmd.StdoutPipe()
	serr, _ := cmd.StderrPipe()

	cmd.Env = os.Environ()

	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error starting the program", err)
	}
	if !quiet {

		go io.Copy(os.Stdout, sout)
		go io.Copy(os.Stdout, serr)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error waiting program")
	}

	return err
}
