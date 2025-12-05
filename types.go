package netbox

import (
	"fmt"
	"strings"
)

type JournalLevel int

// JournalLevels
const (
	Undefined JournalLevel = iota
	InfoLevel
	SuccessLevel
	DangerLevel
	WarningLevel
)

type MonitoredObject struct {
	ID         int64  `json:"id"`
	URL        string `json:"url"`
	ObjectType string `json:"-"`
}

type MonitoringSearchResults struct {
	Count    int               `json:"count"`
	Next     interface{}       `json:"next"`
	Previous interface{}       `json:"previous"`
	Results  []MonitoredObject `json:"results"`
}

// getObjectType returns the full netbox object type for the given model.
// For example, given the type of "device" will return "dcim.device"
func getObjectType(aModel string) string {
	var group string
	switch aModel {
	case "interface":
		fallthrough
	case "location":
		fallthrough
	case "device":
		group = "dcim"
	case "cluster-group":
		aModel = "clustergroup"
		group = "virtualization"
	case "cluster-type":
		aModel = "clustertype"
		group = "virtualization"
	case "vminterface":
		fallthrough
	case "cluster":
		fallthrough
	case "virtualmachine":
		group = "virtualization"
	case "ipaddress":
		group = "ipam"
	case "aggregate":
		group = "ipam"
	case "prefix":
		group = "ipam"
	case "ip-range":
		group = "ipam"
	default:
		return "Invalid"
	}
	return fmt.Sprintf("%s.%s", group, aModel)
}

func getJournalLevel(level JournalLevel) string {
	switch level {
	case Undefined:
		return ""
	case InfoLevel:
		return "info"
	case SuccessLevel:
		return "success"
	case WarningLevel:
		return "warning"
	case DangerLevel:
		return "danger"
	}
	return ""
}

type DeviceOrVM struct {
	AssetTag     *string `json:"asset_tag"`
	Comments     string  `json:"comments"`
	Created      string  `json:"created"`
	CustomFields struct {
		MonitoringID *int `json:"monitoring_id"`
	} `json:"-"`
	CustomFieldsMap map[string]interface{} `json:"custom_fields"`
	Description     string                 `json:"description"`
	DeviceRole      DisplayIDName          `json:"device_role"`
	Display         string                 `json:"display"`
	ID              int                    `json:"id"`
	LastUpdated     string                 `json:"last_updated"`
	Latitude        *float64               `json:"latitude"`
	Longitude       *float64               `json:"longitude"`
	Name            string                 `json:"name"`
	PrimaryIP       PrimaryI               `json:"primary_ip"`
	PrimaryIp4      PrimaryI               `json:"primary_ip4"`
	Rack            struct {
		Display string `json:"display"`
		ID      int    `json:"id"`
		Name    string `json:"name"`
		URL     string `json:"url"`
	} `json:"rack"`
	Role      DisplayIDName `json:"role"`
	Serial    string        `json:"serial"`
	Site      DisplayIDName `json:"site"`
	Status    LabelValue    `json:"status"`
	URL       string        `json:"url"`
	Memory    int           `json:"memory,omitempty"`
	Diskspace int           `json:"disk,omitempty"`
	VCPUs     float32       `json:"vcpus,omitempty"`
}
type DisplayIDName struct {
	Display string `json:"display"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	URL     string `json:"url"`
}
type LabelValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
type PrimaryI struct {
	Address string      `json:"address"`
	Display string      `json:"display"`
	Family  interface{} `json:"family"`
	ID      int         `json:"id"`
	URL     string      `json:"url"`
}

// DeviceVMSearchResults are returned for searches of
// either devices or virtual machines
type DeviceVMSearchResults struct {
	Count    int          `json:"count"`
	Next     *string      `json:"next"`
	Previous *string      `json:"previous"`
	Results  []DeviceOrVM `json:"results"`
}

type MAC struct {
	MacAddress *string `json:"mac_address,omitempty"`
	ID         int     `json:"id,omitempty"`
	URL        string  `json:"url,omitempty"`
}

type Interface struct {
	Bridge interface{} `json:"bridge"`
	Cable  *struct {
		Display string `json:"display"`
		ID      int    `json:"id"`
		Label   string `json:"label"`
		URL     string `json:"url"`
	} `json:"cable"`
	CableEnd                    string                 `json:"cable_end"`
	ConnectedEndpoints          interface{}            `json:"connected_endpoints"`
	ConnectedEndpointsReachable *bool                  `json:"connected_endpoints_reachable"`
	ConnectedEndpointsType      interface{}            `json:"connected_endpoints_type"`
	CountFhrpGroups             int                    `json:"count_fhrp_groups"`
	CountIpaddresses            int                    `json:"count_ipaddresses"`
	Created                     string                 `json:"created"`
	CustomFields                map[string]interface{} `json:"custom_fields"`
	Description                 string                 `json:"description"`
	Device                      DisplayIDName          `json:"device"`
	Display                     string                 `json:"display"`
	Duplex                      *LabelValue            `json:"duplex"`
	Enabled                     bool                   `json:"enabled"`
	ID                          int                    `json:"id"`
	L2vpnTermination            interface{}            `json:"l2vpn_termination"`
	Label                       string                 `json:"label"`
	Lag                         interface{}            `json:"lag"`
	LastUpdated                 string                 `json:"last_updated"`
	LinkPeers                   []struct {
		Cable    int           `json:"cable"`
		Device   DisplayIDName `json:"device"`
		Display  string        `json:"display"`
		ID       int           `json:"id"`
		Name     string        `json:"name"`
		Occupied bool          `json:"_occupied"`
		URL      string        `json:"url"`
	} `json:"link_peers"`
	LinkPeersType      *string       `json:"link_peers_type"`
	PrimaryMAC         *MAC          `json:"primary_mac_address,omitempty"`
	MarkConnected      bool          `json:"mark_connected"`
	MgmtOnly           bool          `json:"mgmt_only"`
	Mode               interface{}   `json:"mode"`
	Module             interface{}   `json:"module"`
	Mtu                interface{}   `json:"mtu"`
	Name               string        `json:"name"`
	Occupied           bool          `json:"_occupied"`
	Parent             interface{}   `json:"parent"`
	PoeMode            interface{}   `json:"poe_mode"`
	PoeType            interface{}   `json:"poe_type"`
	RfChannel          interface{}   `json:"rf_channel"`
	RfChannelFrequency interface{}   `json:"rf_channel_frequency"`
	RfChannelWidth     interface{}   `json:"rf_channel_width"`
	RfRole             interface{}   `json:"rf_role"`
	Speed              *int          `json:"speed"`
	TaggedVlans        []interface{} `json:"tagged_vlans"`
	Tags               []interface{} `json:"tags"`
	TxPower            interface{}   `json:"tx_power"`
	Type               struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"type"`
	URL          string        `json:"url"`
	UntaggedVlan interface{}   `json:"untagged_vlan"`
	Vdcs         []interface{} `json:"vdcs"`
	Vrf          interface{}   `json:"vrf"`
	WirelessLans []interface{} `json:"wireless_lans"`
	WirelessLink interface{}   `json:"wireless_link"`
	Wwn          interface{}   `json:"wwn"`
}

func (i *Interface) GetSpeed() int {
	if i.Speed == nil {
		return 0
	}
	return *i.Speed
}

func (i *Interface) GetDuplex() string {
	if i.Duplex == nil {
		return "auto"
	}
	return i.Duplex.Value
}

func (i *Interface) GetMacAddress() string {
	var mac string
	if i.PrimaryMAC != nil && i.PrimaryMAC.MacAddress != nil {
		mac = *i.PrimaryMAC.MacAddress
	}
	return mac
}

type InterfacesResponse struct {
	Count    int         `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  []Interface `json:"results"`
}

// InterfaceEdit is used to add/update an interface
type InterfaceEdit struct {
	Description string      `json:"description,omitempty"`
	Device      *int        `json:"device,omitempty"`
	VM          *int        `json:"virtual_machine,omitempty"`
	Display     *string     `json:"display,omitempty"`
	Duplex      *string     `json:"duplex,omitempty"`
	Label       *string     `json:"label,omitempty"`
	Lag         interface{} `json:"lag,omitempty"`
	PrimaryMAC  float64     `json:"primary_mac_address,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Speed       *int        `json:"speed,omitempty"`
	Type        *string     `json:"type,omitempty"`
	Parent      *int        `json:"parent,omitempty"`
}

// SetSpeed sets the speed to update.  Returns true
// if the value is changed
func (i *InterfaceEdit) SetSpeed(speed int) bool {
	if speed == 0 {
		return false
	}
	i.Speed = &speed
	return true
}

func (i *InterfaceEdit) SetDuplex(duplex *string) bool {
	newDuplex := "auto"

	if duplex == nil {
		return false
	}
	if *duplex == "unknown" {
		return false
	}
	if strings.HasPrefix(*duplex, "full") {
		newDuplex = "full"
	}
	if strings.HasPrefix(*duplex, "half") {
		newDuplex = "half"
	}

	i.Duplex = &newDuplex
	return true
}

func (i *InterfaceEdit) SetMac(macid float64) bool {
	i.PrimaryMAC = macid
	return true
}

func (i *InterfaceEdit) SetName(name string) bool {
	i.Name = &name
	return true
}

func (i *InterfaceEdit) SetParent(parent int) bool {
	if parent != 0 {
		i.Parent = &parent
		return true
	}
	return false
}
