# Architecture

The firegopher project is made up of three main parts:

1. The **VM Runner**, which prepares the host system and starts the guest VM
2. The **Guest Init System**, which runs inside of the guest VM and starts the user workload
3. A set of customised root filesystems ([Base Images]()) to be used for the guest VM and an **Asset Manager** to install and manage them

![Diagram of the inner workings](assets/diagram.jpg)

## VM Runner
The VM Runner has four main responsibilities:

1. It prepares the root file system for the guest VM
    - by unpacking the user assets into a copy of the chosen [Base Image]()
    - and creating a configuration file that instructs the [Guest Init System]() on what to do with them
2. It prepares the [Firecracker Jail]() 
    - by first creating a directory that will later be used by the Firecracker Jailer as a [CHROOT root directory]()
    - and copying/hardlinking all the assets that are needed to run the guest VM into it
3. It creates and configures the network device needed for the guest VM
4. It starts the jailed Firecracker process and supervises it

## Guest Init System
The Guest Init System is a barebones init system specifically designed to run a single application inside of a VM. At its core it is a go-rewrite of [fly-init-snapshot](https://github.com/superfly/init-snapshot).

The main responsbilities of the guest init system are:

1. reading the configuration passed to it from the VM Runner
2. mounting the root file system
3. mounting all the required [device files]()
4. configuring the network interface
5. dropping root priviledges
6. starting the user workload and supervising it


## Asset Manager
The Asset Manager is currently in the early stages of its development. In its current state it is little more than a glorified download script. 

It downloads the following assets:

1. A Linux Kernel Image
2. A specific version of Firecracker (currently 1.6.0)
3. A modified version of the Ubuntu 22.04 root file system


## Base Images
Currently there is only one officialy supported base image available to be used with firegopher. It is a slgihtly modified version of [Ubuntu 22.04 Minimal](https://cloud-images.ubuntu.com/minimal/releases/jammy/release/).

The base image has been modified by running the following bash script inside a container running [public.ecr.aws/ubuntu/ubuntu:jammy](https://gallery.ecr.aws/ubuntu/ubuntu):

```bash
# Install dependencies to run a basic python-based demo application
apt-get update
apt-get install -y ca-certificates 
apt-get install -y curl
apt-get install -y python3
rm -rf /var/lib/apt/lists/*


# Copy everything we need to the bind-mounted rootfs image file
dirs="bin etc home lib lib64 root sbin usr"
for d in $dirs; do tar c "/$d" | tar x -C $rootfs; done

# Make mountpoints
mkdir -pv $rootfs/{dev,proc,sys,run,tmp,var/lib/systemd}
```

The process of creating this base image is based on the process [outlined by the Firecracker team here](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md#creating-a-rootfs-image).