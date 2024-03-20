package netbox

import (
	"fmt"
)

const (
	ipPath = "/ipam/ip-addresses/"
)

// SearchIP searches for the given IP as an ipaddress.
func (c *Client) SearchIP(ip string) (*IPSearchResults, error) {
	obj := &IPSearchResults{}
	r := c.buildRequest().SetResult(obj)
	url := fmt.Sprintf("%s?address=%s", c.buildURL(ipPath), ip)
	resp, err := r.Get(url)
	if err != nil {
		c.log.Error("Could not find address", "err", err)
		return obj, err
	}
	if err = checkStatus(resp); err != nil {
		c.log.Error("Error returned by netbox", "err", err)
		return obj, err
	}
	return obj, err
}

// SetIPDNS searches for the IP given and updates the DNS address
// with the provided FQDN.  It updates all matching ipaddress
// records where the dnsname is not already set.
func (c *Client) SetIPDNS(ip string, dns string) error {
	obj, err := c.SearchIP(ip)
	if err != nil {
		c.log.Error("Could not find address", "err", err)
		return err
	}
	for _, addr := range obj.Results {
		if addr.DNSName == "" {
			c.UpdateAddress(addr.URL, dns)
		}
	}
	return err
}

// UpdateAddress updates the ipaddress indicated the by the URL
// with the given dns FQDN
func (c *Client) UpdateAddress(url, dns string) {
	data := make(map[string]interface{})
	data["dns_name"] = dns
	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(&obj)
	r.SetBody(data)
	resp, err := r.Patch(url)
	if err != nil {
		c.log.Error("Could not update DNS", "url", url, "err", err)
	}
	if err = checkStatus(resp); err != nil {
		c.log.Error("Error returned by netbox", "url", url, "err", err)
	}
}

// AddIP adds an IP address to Netbox
func (c *Client) AddIP(ipaddress string) error {
	data := make(map[string]interface{})
	data["address"] = ipaddress
	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(&obj)
	r.SetBody(data)
	resp, err := r.Post(c.buildURL(ipPath))
	if err != nil {
		return err
	}
	err = checkStatus(resp)

	return err
}
