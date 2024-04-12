package main

import (
	assetmanager "firegopher.dev/asset-manager"
	"github.com/urfave/cli"
)

var bootstrapCmd = cli.Command{
	Name:    "bootstrap",
	Aliases: []string{"bt"},
	Usage:   "Install firecracker and download base images, init-binary and kernel",
	Action:  bootstrap,
}

func bootstrap(c *cli.Context) {
	assetmanager.InstallFirecracker()
	assetmanager.DownloadAssets()

}
