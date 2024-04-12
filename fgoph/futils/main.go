package futils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func CreateFileWithContent(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {

		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func Mkdir(dir string, mode uint32) error {
	err := syscall.Mkdir(dir, mode)
	if err != nil {
		fmt.Printf("Did not create dir %s: %v\n", dir, err)
		return err
	}
	fmt.Printf("Dir %s created successfully\n", dir)
	return nil
}

func Mount(source string, target string, fstype string, flags uintptr, data string) error {
	err := syscall.Mount(source, target, fstype, flags, data)

	if err != nil {
		fmt.Printf("Error mounting %s: %v\n", target, err)
		return err
	}
	fmt.Printf("%s mounted successfully\n", target)
	return nil
}

func Unmount(target string) error {
	err := syscall.Unmount(target, 0)
	if err != nil {
		fmt.Printf("Error unmounting %s: %v\n", target, err)
		return err
	}
	fmt.Printf("%s unmounted successfully\n", target)
	return nil
}

func CopyFile(src, dst string) (err error) {

	cpCmd := exec.Command("cp", src, dst)

	cpCmd.Env = os.Environ()
	sout, _ := cpCmd.StdoutPipe()
	serr, _ := cpCmd.StderrPipe()
	go io.Copy(os.Stdout, sout)
	go io.Copy(os.Stdout, serr)

	if err := cpCmd.Run(); err != nil {
		log.Fatalf("Failed to create root.ext4 file: %v", err)
		return err
	}

	return err

}

func Sha1Sum(filePath string) (checksum string) {
	cmd := exec.Command("sha1sum", filePath)
	log.Println(cmd)
	cmd.Env = os.Environ()
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	cmd.Dir = path
	// Run the command and get the output
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute command: %s", err)
	}

	// Convert output bytes to a string
	checksum = string(out)
	return checksum[:40]
}

//Figure out wtf happened here

func CopyFileSlow(src, dst string) (err error) {

	fmt.Printf("Copying %s to %s\n", src, dst)
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Printf("There was an error with opening file %s\n", src)
		fmt.Println(err)
		return err
	}
	defer srcFile.Close()

	// Create the destination file for writing
	dstFile, err := os.Create(dst)
	if err != nil {
		fmt.Printf("There was an error with creating file %s\n", dst)
		fmt.Println(err)

		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		fmt.Printf("There was an error with copying to %s\n", dst)
		fmt.Println(err)

		return err
	}

	// The copy was successful. Call Sync to ensure that any buffered data is written to disk.
	return dstFile.Sync()
}

//This was the result of stupidy but it is kinda more fun to use than Walk by go so I dont care
func WalkDirFast(dir string, processFunc func(fileType string, filePath string)) (err error) {
	cmd := exec.Command("find", dir, "-printf", "%y%p\n")
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		processFunc(string(line[0]), line[len(dir)+1:])
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
