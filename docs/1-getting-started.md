# Getting Started

## Prerequisites

Firegopher currently only officially supports [Ubuntu Jammy Jellyfish](https://releases.ubuntu.com/jammy/) as a host operating system. This is mostly due to the fact that firegopher depends on [debugfs >=1.46.5](https://e2fsprogs.sourceforge.net/). 

In general firegopher only works on linux as it uses [Firecracker](https://firecracker-microvm.github.io/) which uses [KVM](https://en.wikipedia.org/wiki/Kernel-based_Virtual_Machine).


## Installing firegopher
To install firegopher run the install script from the website.

    curl -fsSL https://firegopher.dev/install | bash


## Downloading assets and additional dependencies
Firegopher needs a kernel and root filesystem for its VMs and also a version of Firecracker (including Jailer) to run them. 
To make the installation of these dependencies easier firegopher comes with a bootstrap command that installs them.

    sudo fgoph bootstrap


## Run your first VM
You can download an example app.zip file to use for demo purposes

    wget https://firegopher.dev/example-app.zip

To run the example application use

    sudo fgoph run --app="example-app.zip" --exec="python3 app.py"

The command output should tell you the virtual IP that has been assigned to the VM. 

You can try sending a request to the app by running the following command in a new terminal session:

    curl http://172.19.0.2:8000

(The IP might be different depending on what IPs are still available for asignment on your host machine)