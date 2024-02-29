package netbox

import "strings"

// IPfromCIDR takes an IP address in CIDR notation
// and returns just the IP without the mask.
func IPfromCIDR(cidr string) string {
	return strings.Split(cidr, "/")[0]
}

func GetPathForModel(model string) string {
	path := ""

	switch model {
	case "interface":
		path = "/dcim/interfaces"
	case "device":
		path = "/dcim/devices"
	case "ipaddress":
		fallthrough
	case "ip-address":
		path = "/ipam/ip-addresses"
	case "virtualmachine":
		path = "/virtualization/virtual-machines"
	}
	return path
}
