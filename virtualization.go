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

type ClusterTypesResponse struct {
	Count    int           `json:"count"`
	Next     interface{}   `json:"next"`
	Previous interface{}   `json:"previous"`
	Results  []ClusterType `json:"results"`
}

type ClusterType struct {
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

// GetCluster looks up the cluster with the given name in the given group
func (c *Client) GetCluster(group string, name string) (Cluster, error) {
	var cluster Cluster
	results := &ClusterResponse{}
	err := c.search("cluster", results, fmt.Sprintf("group=%s&name=%s", group, name))
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

// AddCluster creates a new cluster in the given group
func (c *Client) AddCluster(group string, name string, clusterType string) (any, error) {
	var cluster Cluster
	cGroup, err := c.GetClusterGroup(group)
	if err != nil {
		c.log.Error("Cannot determine cluster group id", "group", group, "error", err)
		return cluster, err
	}
	cType, err := c.GetClusterType(clusterType)
	if err != nil {
		return cluster, err
	}
	data := make(map[string]interface{})
	data["name"] = name
	data["group"] = cGroup.ID
	data["type"] = cType.ID
	r := c.buildRequest().SetResult(&cluster)
	r.SetBody(data)
	path := GetPathForModel("cluster") + "/"
	resp, err := r.Post(c.buildURL(path))
	if err != nil {
		return cluster, err
	}
	if resp.IsError() {
		c.log.Error("error adding cluster", "cluster", name, "error", err)
		return cluster, errors.New("error adding cluster")
	}
	return cluster, nil
}

// GetOrAddCluster will retrieve the cluster if found and add it if it does
// not exist.
func (c *Client) GetOrAddCluster(group string, name string, clusterType string) (any, error) {
	cluster, err := c.GetCluster(group, name)
	if err == nil {
		return cluster, err
	}
	if errors.Is(err, ErrNotFound) {
		return c.AddCluster(group, name, clusterType)
	}
	return cluster, err
}

// SearchVMs searches  the   virtualmachines
// endpoint for the given args.  Args should be specified as
// key=value (eg. has_primary_ip=true)
func (c *Client) SearchVMs(args ...string) ([]DeviceOrVM, error) {
	return c.performDevVMsearch("virtualmachine", args...)
}

// GetClusterType looks up the type by name
func (c *Client) GetClusterType(name string) (ClusterType, error) {
	var cType ClusterType
	results := &ClusterTypesResponse{}
	args := []string{fmt.Sprintf("slug=%s", Slugify(name))}
	err := c.search("cluster-type", results, args...)
	if err != nil {
		c.log.Error("error finding cluster type", "type", name, "error", err)
		return cType, err
	}
	switch results.Count {
	case 0:
		return cType, ErrNotFound
	case 1:
		return results.Results[0], nil
	}
	return cType, errors.New("too many results returned")
}
