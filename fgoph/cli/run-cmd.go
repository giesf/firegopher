package main

import (
	"strings"

	"firegopher.dev/futils"
	"firegopher.dev/runner"
	runnerconfig "firegopher.dev/runner-config"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "Run a vm",
	Action:  run,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tapDevice, tapDev, tap",
			Usage: "Define the tap device to use for the application networking",
		}, cli.StringFlag{
			Name:  "dataVolume, vol",
			Value: "NONE",
			Usage: "Provide an image to be used as a persistent volume",
		}, cli.StringFlag{
			Name:  "baseImage, os",
			Value: "ubuntu",
			Usage: "Select a base image from /srv/firegopher/*.ext4",
		}, cli.StringFlag{
			Name:  "workingDir, cwd",
			Value: "/app",
			Usage: "The directory to run the entry point command in",
		}, cli.StringFlag{
			Name:  "quiet, q",
			Value: "false",
			Usage: "Disable displaying the log of the vm",
		}, cli.StringFlag{
			Name:  "exec",
			Value: "echo 'Hello, world'",
			Usage: "Set the entrypoint command for the vm",
		},
		cli.StringFlag{
			Name:  "app, f",
			Value: "",
			Usage: "Set the path of the app.zip file to be copied into the vm",
		},
	},
}

func run(c *cli.Context) {

	subargs := strings.Split(c.String("exec"), " ")
	var subcommandArgs []string = subargs[1:]
	var subcommandCmd string = subargs[0]

	tapId, _ := gonanoid.Generate("abcdefghijklmnopqrstuvw123456789", 10)

	fallbackTapDeviceId := "tap-" + tapId

	tapDeviceName := c.String("tapDevice")
	if tapDeviceName == "" {
		tapDeviceName = fallbackTapDeviceId
	}
	appZip := c.String("app")
	dataVolume := c.String("dataVolume")
	baseImage := c.String("baseImage")
	dir := c.String("workingDir")

	quiet := c.Bool("quiet")

	//@TODO make this clean
	codeChecksum := "nocode"
	if appZip != "NONE" {
		codeChecksum = futils.Sha1Sum(appZip)
	}

	//@TODO make this real
	configChecksum, _ := gonanoid.Generate("abcdefghijklmnopqrstuvw123456789", 40)

	runChecksum := configChecksum + codeChecksum

	runConfig := runnerconfig.RunnerConfig{
		CheckSum: runChecksum,
		Quiet:    quiet,
		Workload: runnerconfig.WorkloadConfig{
			AppZip:     appZip,
			DataVolume: dataVolume,
			Cmd:        subcommandCmd,
			Args:       subcommandArgs,
			Dir:        dir,
		},
		Os: runnerconfig.OSConfig{
			BaseImage: baseImage,
		},
		Networking: runnerconfig.NetworkingConfig{
			TapDeviceName: tapDeviceName,
		},
	}
	runner.Runner(runConfig)
}
