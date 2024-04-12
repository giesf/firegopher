# Advanced Usage & Known Limitations

## Networking
The current state of the networking stack should be considered unstable. Every time you run a new microVM firegopher creates a new tap device and assigns it an unused IP from the range `172.19.0.0/16`. 

This is somewhat wasteful as this results in creating a lot of dangling network devices and assigned IPs. You are currently responsible for cleaning them up. 

There is currently no way of predefining a static IP for a microVM.

## Base images
Currently the only officially supported base image is a slightly modified version of [Ubuntu 22.04 Minimal](https://cloud-images.ubuntu.com/minimal/releases/jammy/release/) for demo purposes.

You can create your own base image to be used with firegopher by following the process [outlined by the Firecracker team here](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md#creating-a-rootfs-image) and placing the resulting `imagename.ext4` file in `/srv/firegopher`. Afterwards you can use the base image by running firegopher with the flag `--baseImage=imagename`.

## Persistent storage
Currently you can only run firegopher with a single persistent volume that is always mounted to `/data`.

Right now you are responsible for creating a image file to back that volume and provide it through the dataVolume flag (e.g. `--dateVolume=volume.img`).

## Firegopher should be considered unstable and unfit for production use-cases
In addition to the lacking testing infrastructure **this project currently does not implement the basic security and performance considerations needed to run Firecracker in production.** 

This is due to the fact that it was created as a proof of concept in the context of a university project at [CODE UAS Berlin](https://code.berlin). 

### Firecracker production host recommendations
Currently firegopher does not implement [the production host setup recommendations that the AWS Firecracker team has provided in their git repo](https://github.com/firecracker-microvm/firecracker/blob/main/docs/prod-host-setup.md). 

### Memory usage
Firecracker has the habit of eating up as much ram as it can get its hands on without returning it to the system. [You can read a Hacker News thread discussing this here](https://news.ycombinator.com/item?id=36667792).

An option for dealing with this is [using a memory baloon device with Firecracker](https://github.com/firecracker-microvm/firecracker/blob/main/docs/ballooning.md), which firegopher currently does not do. 