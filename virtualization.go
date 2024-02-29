package netbox

import (
	"errors"
	"fmt"
)

type ClusterGroupResponse struct {
	Count    int            `json:"count"`
	Next     interface{}    `json:"next"`
	Previous interface{}    `json:"previous"`
	Results  []ClusterGroup `json:"results"`
}

type ClusterGroup struct {
	ClusterCount int                    `json:"cluster_count"`
	Created      string                 `json:"created"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Description  string                 `json:"description"`
	Display      string                 `json:"display"`
	ID           int                    `json:"id"`
	LastUpdated  string                 `json:"last_updated"`
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Tags         []interface{}          `json:"tags"`
	URL          string                 `json:"url"`
}

type ClusterResponse struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Cluster   `json:"results"`
}

type Cluster struct {
	Comments     string                 `json:"comments"`
	Created      string                 `json:"created"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Description  string                 `json:"description"`
	DeviceCount  int                    `json:"device_count"`
	Display      string                 `json:"display"`
	Group        DisplayIDName          `json:"group"`
	ID           int                    `json:"id"`
	LastUpdated  string                 `json:"last_updated"`
	Name         string                 `json:"name"`
	Site         DisplayIDName          `json:"site"`
	Status       struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"status"`
	Tags                []interface{} `json:"tags"`
	Tenant              interface{}   `json:"tenant"`
	Type                DisplayIDName `json:"type"`
	URL                 string        `json:"url"`
	VirtualmachineCount int           `json:"virtualmachine_count"`
}

// GetClusterGroup looks up the cluster by name
func (c *Client) GetClusterGroup(name string) (ClusterGroup, error) {
	var group ClusterGroup
	results := &ClusterGroupResponse{}
	err := c.search("cluster-group", results, fmt.Sprintf("name=%s", name))
	if err != nil {
		c.log.Error("error finding cluster group", "group", name, "error", err)
		return group, err
	}
	switch results.Count {
	case 0:
		return group, ErrNotFound
	case 1:
		return results.Results[0], nil
	}
	return group, errors.New("too many results returned")
}

// TODO: Implement AddClusterGroup
func (c *Client) AddClusterGroup(name string) (ClusterGroup, error) {
	group := ClusterGroup{}
	return group, ErrNotImplemented
}

// GetOrAddClusterGroup will retrieve the requested cluster group
// by name and add it if it does not exist
func (c *Client) GetOrAddClusterGroup(name string) (ClusterGroup, error) {
	cluster, err := c.GetClusterGroup(name)
	if err == nil {
		return cluster, err
	}
	if errors.Is(err, ErrNotFound) {
		return c.AddClusterGroup(name)
	}
	return cluster, err
}

// GetCluster looks up the cluster with the given name
func (c *Client) GetCluster(name string) (Cluster, error) {
	var cluster Cluster
	results := &ClusterResponse{}
	err := c.search("cluster", results, fmt.Sprintf("name=%s", name))
	if err != nil {
		c.log.Error("error finding cluster", "cluster", name, "error", err)
		return cluster, err
	}
	switch results.Count {
	case 0:
		return cluster, ErrNotFound
	case 1:
		return results.Results[0], nil
	}
	return cluster, errors.New("too many results returned")
}

// TODO: Implement Addcluster
func (c *Client) AddCluster(name string, clusterType string) (any, error) {
	return nil, ErrNotImplemented
}

// GetOrAddCluster will retrieve the cluster if found and add it if it does
// not exist.
func (c *Client) GetOrAddCluster(name string, clusterType string) (any, error) {
	cluster, err := c.GetCluster(name)
	if err == nil {
		return cluster, err
	}
	if errors.Is(err, ErrNotFound) {
		return c.AddCluster(name, clusterType)
	}
	return cluster, err
}

// SearchVMs searches  the   virtualmachines
// endpoint for the given args.  Args should be specified as
// key=value (eg. has_primary_ip=true)
func (c *Client) SearchVMs(args ...string) ([]DeviceOrVM, error) {
	return c.performDevVMsearch("virtualmachine", args...)
}
