package netbox

import (
	"errors"
	"fmt"
)

// FindInterfaceByName searches Netbox for the given interface name on the requested device
func (c *Client) FindInterfaceByName(netboxType string, netboxDevice int64, ifName string) (intf Interface, err error) {
	intfs, err := c.searchInterfaces(netboxType, netboxDevice, fmt.Sprintf("name=%s", ifName))
	if err != nil {
		return intf, err
	}
	if len(intfs) == 0 {
		return intf, ErrNotFound
	}
	return intfs[0], nil
}

// GetInterfacesforDevices returns all interfaces for the given device.
func (c *Client) GetInterfacesForObject(netboxType string, netboxDevice int64) (intfs []Interface, err error) {
	return c.searchInterfaces(netboxType, netboxDevice)
}

func (c *Client) searchInterfaces(netboxType string, netboxDevice int64, args ...string) (intfs []Interface, err error) {
	var url *string
	id := "device_id"

	model, err := getInterfaceType(netboxType)
	if err != nil {
		return nil, err
	}
	if netboxType == "virtualmachine" {
		id = "virtual_machine_id"
	}
	obj := &InterfacesResponse{}
	r := c.buildRequest().SetResult(obj)
	path := GetPathForModel(model) + "/?" + id + "=%d"
	for _, arg := range args {
		path = fmt.Sprintf("%s&%s", path, arg)
	}
	path = c.buildURL(path, netboxDevice)
	url = &path
	for url != nil {
		resp, err := r.Get(*url)
		if err != nil {
			c.log.Error(fmt.Sprintf("error searching %s", r.URL), "err", err)
			return intfs, err
		}
		if resp.IsError() {
			c.log.Error(fmt.Sprintf("%d searching %s", resp.StatusCode(), r.URL), "err", err)
			return intfs, err
		}
		intfs = append(intfs, obj.Results...)
		url = obj.Next
	}
	return intfs, nil

}

func getInterfaceType(netboxType string) (string, error) {
	model := map[string]string{"device": "interface", "virtualmachine": "vminterface"}
	if netboxType != "device" && netboxType != "virtualmachine" {
		return "nil", errors.New("netboxType must be one of 'device' or 'virtualmachine'")
	}
	return model[netboxType], nil
}

// AddInterface will create a new interface on the given device
func (c *Client) AddInterface(netboxType string, netboxDevice int64, intf InterfaceEdit) (Interface, error) {
	devid := int(netboxDevice)
	newIntf := Interface{}
	intf.Device = &devid
	ifType, err := getInterfaceType(netboxType)
	if err != nil {
		return newIntf, err
	}
	r := c.buildRequest().SetResult(&newIntf).SetBody(intf)

	c.log.Info("adding interface", "body", r.Body)
	resp, err := r.Post(c.buildURL(GetPathForModel(ifType) + "/"))
	if err != nil {
		c.log.Error("error adding interface", "device", netboxDevice, "interface", intf.Name, "error", err)
		return newIntf, err
	}
	if err = checkStatus(resp); err != nil {
		c.log.Error("error checking status", "status", resp.StatusCode(), "error", err)
		return newIntf, err
	}
	c.log.Info("add interface", "interface", *intf.Name, "status", resp.StatusCode(), "url", r.URL)
	return newIntf, nil
}

// UpdateInterface modifies the values of the given interface in Netbox
func (c *Client) UpdateInterface(netboxType string, intfID int64, intf InterfaceEdit) error {
	ifType, err := getInterfaceType(netboxType)
	if err != nil {
		return err
	}
	r := c.buildRequest().SetBody(intf)
	resp, err := r.Patch(c.buildURL(GetPathForModel(ifType)+"/%d/", intfID))
	if err != nil {
		c.log.Error("error updating interface", "interface", intfID, "error", err)
		return err
	}
	if err = checkStatus(resp); err != nil {
		c.log.Error("error checking status", "status", resp.StatusCode(), "error", err)
		return err
	}
	c.log.Info("update interface", "interface", intfID, "status", resp.Status(), "url", r.URL)
	return nil
}
