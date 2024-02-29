package netbox

import (
	"fmt"
)

// FindInterfaceByName searches Netbox for the given interface name on the requested device
func (c *Client) FindInterfaceByName(netboxDevice int64, ifName string) (intf Interface, err error) {
	intfs, err := c.searchInterfaces(netboxDevice, fmt.Sprintf("name=%s", ifName))
	if err != nil {
		return intf, err
	}
	if len(intfs) == 0 {
		return intf, ErrNotFound
	}
	return intfs[0], nil
}

// GetInterfacesforDevices returns all interfaces for the given device.
func (c *Client) GetInterfacesForDevice(netboxDevice int64) (intfs []Interface, err error) {
	return c.searchInterfaces(netboxDevice)
}

func (c *Client) searchInterfaces(netboxDevice int64, args ...string) (intfs []Interface, err error) {
	var url *string
	obj := &InterfacesResponse{}
	r := c.buildRequest().SetResult(obj)
	path := GetPathForModel("interface") + "/?device_id=%d"
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

// AddInterface will create a new interface on the given device
func (c *Client) AddInterface(netboxDevice int64, intf InterfaceEdit) error {
	devid := int(netboxDevice)
	intf.Device = &devid
	r := c.buildRequest().SetBody(intf)
	c.log.Info("adding interface", "body", r.Body)
	resp, err := r.Post(c.buildURL(GetPathForModel("interface") + "/"))
	if err != nil {
		c.log.Error("error adding interface", "device", netboxDevice, "interface", intf.Name, "error", err)
		return err
	}
	if err = checkStatus(resp); err != nil {
		c.log.Error("error checking status", "status", resp.StatusCode(), "error", err)
		return err
	}
	c.log.Info("add interface", "interface", *intf.Name, "status", resp.StatusCode(), "url", r.URL)
	return nil
}

// UpdateInterface modifies the values of the given interface in Netbox
func (c *Client) UpdateInterface(intfID int64, intf InterfaceEdit) error {
	r := c.buildRequest().SetBody(intf)
	resp, err := r.Patch(c.buildURL(GetPathForModel("interface")+"/%d/", intfID))
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
