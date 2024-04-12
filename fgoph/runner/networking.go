package runner

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
)

/*
This contains scripts to configure the host-side of the networking set-up
Most of this was written with the help of ChatGPT4 as @giesf has no fucking clue about networking
The logic is basically:
- Create a tap device
- Disable ipv6 on it
- Asign a subnet
- Forward traffic to that tap device

This has not been audited, tread carefully.
*/

//@TODO get this audited by someone who knows their shit

func setupTapDeviceAndIpForwarding(tapName, ifName string) (net.IP, net.IP, *net.IPNet, error) {

	existingSubnets, subnetCheckErr := getExistingSubnets()
	if subnetCheckErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to check for existing subnets: %w", subnetCheckErr)
	}

	freeSubnet, freeSubnetFindErr := findFreeSubnet(existingSubnets)
	if freeSubnetFindErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to find free subnet: %w", freeSubnetFindErr)
	}

	gatewayIP, subnet, err := net.ParseCIDR(freeSubnet)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse free subnet: %w", err)
	}
	fmt.Println(gatewayIP, subnet)

	// Ensure the TAP device exists
	link, deviceErr := ensureTapDeviceExists(tapName)
	if deviceErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to ensure TAP device exists: %w", deviceErr)
	}
	ipErr := configureForwarding(link, tapName, freeSubnet, ifName)
	if ipErr != nil {
		fmt.Println(ipErr)
	}

	guestIP := nextIP(gatewayIP)

	return gatewayIP, guestIP, subnet, nil
}

func nextIP(ip net.IP) net.IP {
	// Convert IP to 4-byte representation
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)
	newIP = newIP.To4()

	for i := len(newIP) - 1; i >= 0; i-- {
		// Increment the last byte
		newIP[i]++
		if newIP[i] != 0 {
			break // No carry, so stop
		}
		// If byte is 0, there was a carry so continue to next byte
	}
	return newIP
}

func getExistingSubnets() ([]string, error) {
	// Command to list IP addresses of TAP devices (simplified example)
	cmd := "ip -o addr show | grep tap | awk '{print $4}'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}

	// Parse output to extract subnets
	subnets := strings.Split(strings.TrimSpace(string(out)), "\n")
	return subnets, nil
}
func findFreeSubnet(existingSubnets []string) (string, error) {
	allExistingSubnets := strings.Join(existingSubnets, "\n")

	// Loop through the entire 172.19.0.0/16 range in /30 blocks
	for secondOctet := 0; secondOctet <= 255; secondOctet++ {
		for thirdOctet := 1; thirdOctet <= 255; thirdOctet += 4 { // Increment by 4 for /30 blocks
			subnet := fmt.Sprintf("172.19.%d.%d/30", secondOctet, thirdOctet)
			if !strings.Contains(allExistingSubnets, subnet) {
				return subnet, nil
			}
		}
	}

	return "", fmt.Errorf("no free subnet found")
}

// ensureTapDeviceExists checks for a TAP device by name, creates it if it does not exist, and brings it up.
// It returns the TAP device as a netlink.Link and any error encountered.
func ensureTapDeviceExists(tapName string) (netlink.Link, error) {
	fmt.Printf("Ensuring %s exists and works properly\n", tapName)
	// Attempt to find the TAP device by name
	link, err := netlink.LinkByName(tapName)
	if err != nil {
		if err.Error() == "Link not found" {
			// TAP device does not exist, so create it
			cmd := exec.Command("sudo", "ip", "tuntap", "add", "dev", tapName, "mode", "tap")
			if err := cmd.Run(); err != nil {
				return nil, fmt.Errorf("tap device could not be created: %w", err)
			}

			upCmd := exec.Command("sudo", "ip", "link", "set", tapName, "up")
			if err := upCmd.Run(); err != nil {
				return nil, fmt.Errorf("tap device failed to change state to up: %w", err)
			}

			disableIPv6Cmd := exec.Command("sudo", "sysctl", "-w", "net.ipv6.conf."+tapName+".disable_ipv6=1")
			disableIPv6Err := disableIPv6Cmd.Run()
			if disableIPv6Err != nil {
				fmt.Println("IPv6 was not disabled")
			}

			link, err = netlink.LinkByName(tapName)
		} else {
			// Some other error occurred
			return nil, fmt.Errorf("error checking TAP device: %w", err)
		}
	}

	// TAP device exists, return the existing device
	return link, err
}

func configureForwarding(link netlink.Link, tapName, tapIP, ifName string) error {

	// Add IP address to TAP device
	addr, err := netlink.ParseAddr(tapIP)
	if err != nil {
		return fmt.Errorf("failed to parse IP address: %w", err)
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("failed to add IP address to TAP device: %w", err)
	}

	// Optionally, show the TAP device address for verification (not required for functionality)
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to list TAP device addresses: %w", err)
	}
	for _, a := range addrs {
		fmt.Println(a.IPNet.String())
	}

	// Set up IP forwarding
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Set up NAT masquerading with iptables
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("failed to initialize iptables: %w", err)
	}

	if err := ipt.AppendUnique("nat", "POSTROUTING", "-o", ifName, "-j", "MASQUERADE"); err != nil {
		return fmt.Errorf("failed to add iptables masquerade rule: %w", err)
	}

	if err := ipt.AppendUnique("filter", "FORWARD", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to add iptables forwarding rule for established connections: %w", err)
	}

	if err := ipt.AppendUnique("filter", "FORWARD", "-i", tapName, "-o", ifName, "-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to add iptables forwarding rule for TAP device: %w", err)
	}

	fmt.Println("Network setup complete.")
	return nil
}

func generateMAC(prefix string) (string, error) {
	buf := make([]byte, 4) // Generate 4 random bytes for the remaining address
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	// Construct MAC address using prefix and random bytes
	// Ensures locally administered and unicast address
	mac := fmt.Sprintf("%s:%02x:%02x:%02x:%02x", prefix, buf[0], buf[1], buf[2], buf[3])

	return mac, nil
}
