package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"firegopher.dev/futils"
	guestconfig "firegopher.dev/guest-config"

	"github.com/vishvananda/netlink"
)

/*
	Runs as PID 1 in the microVM to mount all needed devices and setup the networking before starting the guest app
	Mostly a GO port of https://github.com/superfly/init-snapshot
*/

func debug(s string, a ...any) {
	fmt.Printf(s+"\n", a...)
}

func main() {

	config, err := guestconfig.ReadConfigFromFile("/firegopher/run.toml")

	if err != nil {
		fmt.Printf("Err reading run.toml")
		os.Exit(1)
	}

	debug("Mounting /dev")
	futils.Mkdir("/dev", 0o755)
	futils.Mount("devtmpfs", "/dev", "devtmpfs", syscall.MS_NOSUID, "mode=0755")

	debug("Preparing newroot...")
	futils.Mkdir("/newroot", 0o755)

	var rootDevice string = "/dev/vdb"

	debug("Mount newroot fs")
	futils.Mount(rootDevice, "/newroot", "ext4", syscall.MS_RELATIME, "")

	debug("Mounting (move) /dev")
	futils.Mkdir("/newroot/dev", 0o755)
	futils.Mount("/dev", "/newroot/dev", "", syscall.MS_MOVE, "")

	debug("Removing /firegopher to save space")
	os.RemoveAll("/firegopher")

	debug("Switching to new root")
	os.Chdir("/newroot")
	futils.Mount(".", "/", "", syscall.MS_MOVE, "")
	syscall.Chroot(".")
	os.Chdir("/")

	debug("Mounting /dev/pts")
	futils.Mkdir("/dev/pts", 0o755)
	var devptsMountFlags uintptr = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NOATIME
	devptsMountData := "mode=0620,gid=5,ptmxmode=666"
	futils.Mount("devpts", "/dev/pts", "devpts", devptsMountFlags, devptsMountData)

	var commonMountFlags uintptr = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID

	debug("Mounting /dev/mqueue")
	futils.Mkdir("/dev/mqueue", 0o755)
	futils.Mount("mqueue", "/dev/mqueue", "mqueue", commonMountFlags, "")

	debug("Mounting /dev/hugepages")
	futils.Mkdir("/dev/hugepages", 0o755)
	futils.Mount("hugetlbfs", "/dev/hugepages", "hugetlbfs", syscall.MS_RELATIME, "pagesize=2M")

	debug("Mounting /proc")
	futils.Mkdir("/proc", 0o555)
	futils.Mount("proc", "/proc", "proc", commonMountFlags, "")
	futils.Mount("binfmt_misc", "/proc/sys/fs/binfmt_misc", "binfmt_misc", commonMountFlags|syscall.MS_RELATIME, "")

	debug("Mounting /sys")
	futils.Mkdir("/sys", 0o555)
	futils.Mount("sys", "/sys", "sysfs", commonMountFlags, "")

	debug("Mounting /run")
	futils.Mkdir("/run", 0o755)
	futils.Mount("run", "/run", "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, "mode=0755")
	futils.Mkdir("/run/lock", 0o777)

	syscall.Symlink("/proc/self/fd", "/dev/fd")
	syscall.Symlink("/proc/self/fd/0", "/dev/stdin")
	syscall.Symlink("/proc/self/fd/1", "/dev/stdout")
	syscall.Symlink("/proc/self/fd/2", "/dev/stderr")

	debug("Creating dir /root...")
	futils.Mkdir("/root", 0o700)

	debug("Mounting cgroup")
	var cgroupMountFlags uintptr = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID
	futils.Mount("tmpfs", "/sys/fs/cgroup", "tmpfs", cgroupMountFlags, "mode=755")

	var cgroupCommonMountFlags uintptr = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_RELATIME

	debug("Mounting cgroup2")
	futils.Mkdir("/sys/fs/cgroup/unified", 0o555)
	futils.Mount("cgroup2", "/sys/fs/cgroup/unified", "cgroup2", cgroupCommonMountFlags, "nsdelegate")

	cgroupSubsystems := []string{
		"net_cls,net_prio",
		"hugetlb",
		"pids",
		"freezer",
		"cpuacct",
		"devices",
		"blkio",
		"memory",
		"perf_event",
		"cpuset",
	}

	for _, subsystem := range cgroupSubsystems {
		debug("Mounting /sys/fs/cgroup/" + subsystem)
		futils.Mkdir("/sys/fs/cgroup/"+subsystem, 0o555)
		futils.Mount("cgroup", "/sys/fs/cgroup/"+subsystem, "cgroup", cgroupCommonMountFlags, subsystem)
	}

	//    rlimit::setrlimit(rlimit::Resource::NOFILE, 10240, 10240).ok();
	uid := config.Security.User
	gid := config.Security.Group

	debug("Found user %d and group %d", uid, gid)

	//@TODO Set env variables
	os.Setenv("FOO", "1")

	debug("Trying to mount data volume...")
	var userVolumeDevice string = "/dev/vdc"

	futils.Mkdir("/data", 0o755)
	userVolumeErr := futils.Mount(userVolumeDevice, "/data", "ext4", syscall.MS_RELATIME, "")
	if userVolumeErr != nil {
		debug("User volume not mounted.")
	}
	os.Chown("/data", int(uid), int(gid))
	//@TODO mount user volumes
	//@TODO chown the volumes

	debug("Create /etc")
	futils.Mkdir("/etc", 0o755)

	debug("Writing hostname %s to /etc/hostname", config.Etc.Hostname)
	futils.CreateFileWithContent("/etc/hostname", config.Etc.Hostname)

	//@TODO populate /etc/hosts

	fmt.Printf("%s", config.Etc.Nameservers)
	var resolvContent = strings.Join(mapStrings(config.Etc.Nameservers, func(n string) string {
		return "nameserver " + n
	}), "\n")
	debug("Writing nameservers (%s) to /etc/resolv.conf", resolvContent)
	futils.CreateFileWithContent("/etc/resolv.conf", resolvContent)

	//@TODO Checksum offloading

	debug("Setting lo to up")
	lo, _ := netlink.LinkByName("lo")
	netlink.LinkSetUp(lo)

	debug("Setting eth0 to up")
	eth0, _ := netlink.LinkByName("eth0")
	netlink.LinkSetUp(eth0)

	debug("Configuring ip %s/%s", config.Ip.Ip, config.Ip.Mask)

	addr, err := netlink.ParseAddr(config.Ip.Ip + "/" + config.Ip.Mask)
	if err != nil {
		fmt.Println("Error parsing addr")
		panic(err)
	}
	err = netlink.AddrAdd(eth0, addr)
	if err != nil {
		fmt.Println("Error adding addr")

		panic(err)
	}

	debug("Setting gateway IP to %s", config.Ip.Gateway)
	gateway := net.ParseIP(config.Ip.Gateway)
	if gateway == nil {
		fmt.Println("Invalid gateway IP")
		return
	}

	// Parse the default destination network
	_, defaultDst, _ := net.ParseCIDR("0.0.0.0/0")

	// Set up the route struct with the gateway
	route := netlink.Route{
		Dst:       defaultDst,
		LinkIndex: eth0.Attrs().Index,
		Gw:        gateway,
	}
	debug("Adding route %s", route)

	netlink.RouteAdd(&route)
	//@TODO set ips

	debug("Setting path")
	err = os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	if err != nil {
		panic(err)
	}

	debug("Running cmd %s in dir %s", config.Workload.Cmd, config.Workload.Dir)
	debug("\n\n#############################\n")
	debug("IP: %s", config.Ip.Ip)
	debug("\n#############################")

	debug("\n\n############# START OF USER COMMAND EXECUTION #############\n")
	execute(config.Workload.Cmd, config.Workload.Args, config.Workload.Dir, uid, gid, false)
	debug("\n\n############# END OF USER COMMAND EXECUTION #############\n")

	execute("sync", []string{}, config.Workload.Dir, uid, gid, true)

}
