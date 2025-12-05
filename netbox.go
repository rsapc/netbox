package netbox

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/go-resty/resty/v2"
	"github.com/rsapc/hookcmd/models"
)

var ErrNotFound = errors.New("the requested object was not found")
var ErrNotImplemented = errors.New("not implemented")

const (
	updateSitePath = "/dcim/sites/{id}/"
	addSitePath    = "/dcim/sites/"
	tenantPath     = "/tenancy/tenants/"
)

var slugregex *regexp.Regexp
var embedStartRegex *regexp.Regexp
var (
	apcTag      = Tag{Name: "APC", Slug: "apc"}
	jobberTag   = Tag{Name: "Jobber-Imported", Slug: "jobber"}
	customerTag = Tag{Name: "Customer", Slug: "customer"}

	customerGroup = &Group{Name: "Customer", Slug: "customer", Tags: []Tag{customerTag, apcTag, jobberTag}}
)

var addresses = make(map[string]float64)

type Tag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Tenant struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Tags        []Tag  `json:"tags"`
}

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Tags        []Tag  `json:"tags"`
}

type SearchResults struct {
	Count    int `json:"count"`
	Next     any `json:"next"`
	Previous any `json:"previous"`
	Results  []struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Display string `json:"display"`
		Name    string `json:"name"`
		Slug    string `json:"slug"`
		Status  struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"status"`
		Region struct {
			ID      int    `json:"id"`
			URL     string `json:"url"`
			Display string `json:"display"`
			Name    string `json:"name"`
			Slug    string `json:"slug"`
			Depth   int    `json:"_depth"`
		} `json:"region"`
		Group struct {
			ID      int    `json:"id"`
			URL     string `json:"url"`
			Display string `json:"display"`
			Name    string `json:"name"`
			Slug    string `json:"slug"`
			Depth   int    `json:"_depth"`
		} `json:"group"`
		Tenant struct {
			ID      int    `json:"id"`
			URL     string `json:"url"`
			Display string `json:"display"`
			Name    string `json:"name"`
			Slug    string `json:"slug"`
		} `json:"tenant"`
		Facility        string `json:"facility"`
		TimeZone        any    `json:"time_zone"`
		Description     string `json:"description"`
		PhysicalAddress string `json:"physical_address"`
		ShippingAddress string `json:"shipping_address"`
		Latitude        any    `json:"latitude"`
		Longitude       any    `json:"longitude"`
		Comments        string `json:"comments"`
		Asns            []any  `json:"asns"`
		Tags            []struct {
			ID      int    `json:"id"`
			URL     string `json:"url"`
			Display string `json:"display"`
			Name    string `json:"name"`
			Slug    string `json:"slug"`
			Color   string `json:"color"`
		} `json:"tags"`
		CustomFields struct {
			CommencementDate any `json:"commencement_date"`
			InitialTerm      any `json:"initial_term"`
			Mrc              any `json:"mrc"`
			VendorID         any `json:"vendor_id"`
			WirelessCoverage any `json:"wireless_coverage"`
			Accesscode       any `json:"accesscode"`
			AccessNotes      any `json:"access_notes"`
			ContactEmail     any `json:"contact_email"`
			ContactName      any `json:"contact_name"`
			ContactPhone     any `json:"contact_phone"`
		} `json:"custom_fields"`
		Created             time.Time `json:"created"`
		LastUpdated         time.Time `json:"last_updated"`
		CircuitCount        int       `json:"circuit_count"`
		DeviceCount         int       `json:"device_count"`
		PrefixCount         int       `json:"prefix_count"`
		RackCount           int       `json:"rack_count"`
		VirtualmachineCount int       `json:"virtualmachine_count"`
		VlanCount           int       `json:"vlan_count"`
	} `json:"results"`
}

type IPSearchResults struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []IP        `json:"results"`
}

type IP struct {
	Address        string `json:"address"`
	AssignedObject struct {
		Cable  interface{} `json:"cable"`
		Device struct {
			Display string `json:"display"`
			ID      int    `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
		} `json:"device"`
		Display  string `json:"display"`
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Occupied bool   `json:"_occupied"`
		URL      string `json:"url"`
	} `json:"assigned_object"`
	AssignedObjectID   int    `json:"assigned_object_id"`
	AssignedObjectType string `json:"assigned_object_type"`
	Comments           string `json:"comments"`
	Created            string `json:"created"`
	CustomFields       struct {
	} `json:"custom_fields"`
	DNSName     string `json:"dns_name"`
	Description string `json:"description"`
	Display     string `json:"display"`
	Family      struct {
		Label string `json:"label"`
		Value int    `json:"value"`
	} `json:"family"`
	ID          int           `json:"id"`
	LastUpdated string        `json:"last_updated"`
	NatInside   interface{}   `json:"nat_inside"`
	NatOutside  []interface{} `json:"nat_outside"`
	Role        interface{}   `json:"role"`
	Status      struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"status"`
	Tags []struct {
		Color   string `json:"color"`
		Display string `json:"display"`
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Slug    string `json:"slug"`
		URL     string `json:"url"`
	} `json:"tags"`
	Tenant interface{} `json:"tenant"`
	URL    string      `json:"url"`
	Vrf    interface{} `json:"vrf"`
}

func init() {
	var err error
	slugregex, err = regexp.Compile("[^a-z0-9-_]+")
	if err != nil {
		log.Fatalf("Could not compile slug regex: %v", err)
	}
	embedStartRegex = regexp.MustCompile(
		`(?m:^ *)<!--\s*jobber:import:start\s*-->(?s:.*?)<!--\s*jobber:import:end\s*-->(?m:\s*?$)`,
	)
}

type Client struct {
	client  *resty.Client
	token   string
	baseURL string
	log     models.Logger
}

func NewClient(baseURL string, token string, logger models.Logger) *Client {
	c := &Client{}
	c.client = resty.New()
	c.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	c.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))
	c.baseURL = baseURL
	c.token = token
	c.log = logger
	if log, ok := logger.(*slog.Logger); ok {
		c.log = log.With("service", "netbox")
	}

	return c
}

func (c *Client) buildRequest() *resty.Request {
	return c.client.NewRequest().SetAuthScheme("Token").SetAuthToken(c.token)
}

func (c *Client) buildURL(path string, args ...any) string {
	urlPath := fmt.Sprintf(path, args...)
	return fmt.Sprintf("%s/api%s", c.baseURL, urlPath)
}

func checkStatus(resp *resty.Response) error {
	if resp.IsError() {
		return fmt.Errorf("invalid response to %s %s: [%d] %v %v", resp.Request.Method, resp.Request.URL, resp.StatusCode(), resp.Error(), string(resp.Body()))
	}
	return nil
}

func (c *Client) GetSite(id int) (interface{}, error) {
	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(obj)
	url := fmt.Sprintf("%s/%d", c.buildURL("/dcim/sites"), id)
	resp, err := r.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Result(), checkStatus(resp)
}

func (c *Client) checkSite(param string, name string) (found bool, id int) {
	obj := &SearchResults{}
	r := c.buildRequest().SetResult(obj).SetQueryParam(param, name)
	resp, err := r.Get(c.buildURL("/dcim/sites/"))
	if err != nil {
		return false, 0
	}
	if checkStatus(resp) != nil {
		return false, 0
	}
	resp.Result()
	if obj.Count > 0 {
		site := obj.Results[0]
		return true, site.ID
	}
	return false, 0
}

// FindMonitoredObject searches for the device or VM that has the requested monitoring_id custom field.
func (c *Client) FindMonitoredObject(monitoringID int) (objectType string, objectID int64, err error) {
	obj, err := c.FindMonitoredDevice(monitoringID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return "device", -1, err
		}
	} else {
		return "device", obj.ID, nil
	}
	obj, err = c.FindMonitoredVM(monitoringID)
	return "virtualmachine", obj.ID, err
}

// FindMonitoredDevice searches devices for the given monitoring_id custom field
func (c *Client) FindMonitoredDevice(monitoringID int) (object MonitoredObject, err error) {
	return c.searchMonitoredID(monitoringID, "device")
}

// FindMonitoredVM searches virtual machines for the given monitoring_id custom field
func (c *Client) FindMonitoredVM(monitoringID int) (object MonitoredObject, err error) {
	return c.searchMonitoredID(monitoringID, "virtualmachine")
}

func (c *Client) searchMonitoredID(monitoringID int, objectType string) (object MonitoredObject, err error) {
	path := GetPathForModel(objectType)
	obj := &MonitoringSearchResults{}
	r := c.buildRequest().SetResult(obj)
	url := c.buildURL(fmt.Sprintf("%s/?cf_monitoring_id=%d", path, monitoringID))
	resp, err := r.Get(url)
	if err != nil {
		c.log.Error(fmt.Sprintf("error searching %s", r.URL), "err", err)
		return object, err
	}
	if resp.IsError() {
		c.log.Error(fmt.Sprintf("%d searching %s", resp.StatusCode(), r.URL), "err", err)
		return object, errors.New(string(resp.Body()))
	}
	if obj.Count == 0 {
		return object, ErrNotFound
	}
	if obj.Count > 1 {
		return object, fmt.Errorf("too many objects found: %d", obj.Count)
	}
	object = obj.Results[0]
	object.ObjectType = objectType
	return object, nil
}

// GetDeviceOrVMbyType will return a map representing the object provided by objectType and objectID
func (c *Client) GetDeviceOrVMbyType(objectType string, objectID int64) (obj DeviceOrVM, err error) {
	path := GetPathForModel(objectType)
	if path == "" {
		c.log.Error("could not determine the path for model %s", objectType)
		return obj, fmt.Errorf("could not determine the path for model %s", objectType)
	}
	url := c.buildURL(path+"/%d/", objectID)
	return c.GetDeviceOrVM(url)
}

// GetObject returns a map representing a Netbox Object, retrieved from
// the given URL
func (c *Client) GetDeviceOrVM(url string) (DeviceOrVM, error) {
	obj := DeviceOrVM{}
	r := c.buildRequest().SetResult(&obj)
	resp, err := r.Get(url)
	if err != nil {
		c.log.Error(fmt.Sprintf("error searching %s", r.URL), "err", err)
		return obj, err
	}
	if resp.IsError() {
		c.log.Error(fmt.Sprintf("%d searching %s", resp.StatusCode(), r.URL), "err", err)
		return obj, fmt.Errorf("%s: %s", resp.Error(), resp.Body())
	}
	setDeviceCustomFields(&obj)
	return obj, err
}

func (c *Client) AddSite(data map[string]interface{}) error {
	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(&obj)
	r.SetBody(data)
	resp, err := r.Post(c.buildURL(addSitePath))
	if err != nil {
		return err
	}
	log.Printf("added site: %v %s", obj["id"], data["name"])
	err = checkStatus(resp)

	return err
}

// SetMonitoringID sets the monitoring_id custom field on the given object/id
func (c *Client) SetMonitoringID(model string, modelID int64, devid int) error {
	err := c.UpdateCustomFieldOnModel(model, modelID, "monitoring_id", devid)
	if err != nil {
		c.log.Error(err.Error())
		c.AddJournalEntry(model, modelID, WarningLevel, fmt.Sprintf("failed to add monitoring_id: %d", devid))
		return err
	} else {
		msg := fmt.Sprintf("added monitoring_id %d to %s %d", devid, model, modelID)
		c.AddJournalEntry(model, modelID, SuccessLevel, msg)
	}
	return err
}

// UpdateObjectWithMap takes an object and updates it
func (c *Client) UpdateObject(model string, modelID int64, payload any) error {
	path := GetPathForModel(model)
	if path == "" {
		c.log.Error("could not determine the path for model %s", model)
		return fmt.Errorf("could not determine the path for model %s", model)
	}
	path = fmt.Sprintf("%s/%d/", path, modelID)
	return c.UpdateObjectByURL(c.buildURL(path), payload)
}

// UpdateObjectWithMap takes an object and updates it
func (c *Client) UpdateObjectWithMap(model string, modelID int64, payload map[string]interface{}) error {
	path := GetPathForModel(model)
	if path == "" {
		c.log.Error("could not determine the path for model %s", model)
		return fmt.Errorf("could not determine the path for model %s", model)
	}
	path = fmt.Sprintf("%s/%d/", path, modelID)
	return c.UpdateObjectByURL(c.buildURL(path), payload)
}

func (c *Client) UpdateObjectByURL(url string, payload any) error {
	c.log.Debug(fmt.Sprintf("Updating %s", url))
	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(&obj)
	r.SetBody(payload)
	resp, err := r.Patch(url)
	if err != nil {
		c.log.Warn(err.Error())
		return err
	}
	if resp.IsError() {
		c.log.Error(fmt.Sprintf("invalid response from server: %d: %v", resp.StatusCode(), resp.Error()), "url", r.URL, "body", resp.Body())
		return fmt.Errorf("netbox returned %d: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

// UpdateObjectWithMap takes an object and updates it
func (c *Client) AddObject(model string, payload any) (map[string]interface{}, error) {
	path := GetPathForModel(model)
	if path == "" {
		c.log.Error("could not determine the path for model %s", model)
		return nil, fmt.Errorf("could not determine the path for model %s", model)
	}
	return c.AddObjectByURL(c.buildURL(path), payload)
}
func (c *Client) AddObjectByURL(url string, payload any) (map[string]interface{}, error) {
	c.log.Debug(fmt.Sprintf("Adding %s", url))
	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(&obj)
	r.SetBody(payload)
	resp, err := r.Post(url)
	if err != nil {
		c.log.Warn(err.Error())
		return obj, err
	}
	if resp.IsError() {
		c.log.Error(fmt.Sprintf("invalid response from server: %d: %v", resp.StatusCode(), resp.Error()), "url", r.URL, "body", resp.Body())
		return obj, fmt.Errorf("netbox returned %d: %s", resp.StatusCode(), resp.Body())
	}
	return obj, nil
}

// DeleteObjectByURL will send a DELETE command to the provided URL
func (c *Client) DeleteObjectByURL(url string) error {
	c.log.Debug(fmt.Sprintf("Deleting %s", url))
	r := c.buildRequest()
	resp, err := r.Delete(url)
	if err != nil {
		c.log.Warn(err.Error())
		return err
	}
	if resp.IsError() {
		c.log.Error(fmt.Sprintf("invalid response from server: %d: %v", resp.StatusCode(), resp.Error()), "url", r.URL, "body", resp.Body())
		return fmt.Errorf("netbox returned %d: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

// Slugify takes a string and converts it to a slug by:
// 1. Converting to lowercase.
// 2. Removing characters that aren’t alphanumerics, underscores, hyphens, or whitespace.
// 3. Removing leading and trailing whitespace.
// 4. Replacing any whitespace or repeated dashes with single dashes.
func Slugify(input string) string {
	output := strings.ToLower(input)
	output = strings.TrimSpace(output)
	output = strings.ReplaceAll(output, " ", "-")
	output = slugregex.ReplaceAllString(output, "")

	output = strings.ReplaceAll(output, "--", "-")
	return output
}

func replaceComments(comments string, text string) string {
	embedText := fmt.Sprintf("<!-- config:embed:start -->\n\n%s\n\n<!-- config:embed:end -->", text)

	var replacements int
	os.ReadFile("adsf")
	data := string(embedStartRegex.ReplaceAllFunc([]byte(comments), func(_ []byte) []byte {
		replacements++
		return []byte(embedText)
	}))

	if replacements == 0 {
		log.Printf("no embed markers found. Appending documentation to the end of the file instead")
		return fmt.Sprintf("%s\n\n%s", string(data), text)
	}

	return string(data)
}

func (c *Client) checkGroup(group *Group) {
	obj := &SearchResults{}
	r := c.buildRequest().SetResult(obj).SetQueryParam("slug", group.Slug)
	resp, err := r.Get(c.buildURL("/dcim/site-groups/"))
	if err != nil {
		log.Printf("error searching site-groups: %v\n", err)
		return
	}
	if err = checkStatus(resp); err != nil {
		log.Printf("error checking status: %v", err)
		return
	}
	resp.Result()
	if obj.Count == 0 {
		c.addGroup(group)
	} else {
		group.ID = obj.Results[0].ID
	}
}

func (c *Client) addGroup(group *Group) {
	r := c.buildRequest().SetResult(group).SetBody(group)
	resp, err := r.Post(c.buildURL("/dcim/site-groups/"))
	if err != nil {
		log.Fatalf("error adding group %s: %v\n", group.Name, err)
	}
	if err = checkStatus(resp); err != nil {
		log.Fatalf("error checking status: %v", err)
	}
}

func (c *Client) checkTag(tag Tag) {
	obj := &SearchResults{}
	r := c.buildRequest().SetResult(obj).SetQueryParam("slug", tag.Slug)
	resp, err := r.Get(c.buildURL("/extras/tags/"))
	if err != nil {
		log.Printf("error searching tags: %v\n", err)
		return
	}
	if err = checkStatus(resp); err != nil {
		log.Printf("error checking status: %v", err)
		return
	}
	resp.Result()
	if obj.Count == 0 {
		c.AddTag(tag.Name, tag.Slug)
	}
}

// AddTag creates a new tag in Netbox
func (c *Client) AddTag(name string, slug string) {
	data := make(map[string]interface{})
	data["name"] = name
	data["slug"] = slug
	r := c.buildRequest()
	r.SetBody(data)
	resp, err := r.Post(c.buildURL("/extras/tags/"))
	if err != nil {
		log.Printf("error adding tag: %v\n", err)
		return
	}
	if err = checkStatus(resp); err != nil {
		log.Printf("error checking status: %v", err)
		return
	}
	log.Printf("added tag: %s", name)
}

// AddJournalEntry adds a new journal entry to a location
func (c *Client) AddJournalEntry(model string, modelID int64, level JournalLevel, comments string, args ...any) error {
	data := make(map[string]interface{})
	data["assigned_object_type"] = getObjectType(model)
	data["assigned_object_id"] = modelID
	data["comments"] = fmt.Sprintf(comments, args...)
	levelStr := getJournalLevel(level)
	if levelStr != "" {
		data["kind"] = levelStr
	}

	r := c.buildRequest()
	r.SetBody(data)
	resp, err := r.Post(c.buildURL("/extras/journal-entries/"))
	if err != nil {
		return err
	}
	return checkStatus(resp)
}

func (c *Client) AddLocation(site float64, row map[string]string) error {
	data := make(map[string]interface{})
	data["name"] = row["Service Street 1"]
	data["slug"] = Slugify(fmt.Sprint(row["Service Street 1"]))
	data["site"] = site
	data["status"] = "active"
	data["tags"] = []Tag{apcTag, jobberTag}

	obj := make(map[string]interface{})
	r := c.buildRequest().SetResult(&obj)
	r.SetBody(data)
	resp, err := r.Post(c.buildURL("/dcim/locations/"))
	if err != nil {
		return err
	}
	if err = checkStatus(resp); err != nil {
		return err
	}
	log.Printf("added location %v %s\n", obj["id"], obj["name"])
	return c.AddJournalEntry("location", obj["id"].(int64), InfoLevel, row["comments"])
}

// GetOrAddTenant retrieves the named tenant, or creates it if
// it does not exist
func (c *Client) GetOrAddTenant(name string) (*Tenant, error) {
	obj := &SearchResults{}
	tenant := &Tenant{}
	r := c.buildRequest().SetResult(obj).SetQueryParam("name", name)
	resp, err := r.Get(c.buildURL(tenantPath))
	if err != nil {
		log.Printf("error searching tenants: %v\n", err)
		return nil, err
	}
	if err = checkStatus(resp); err != nil {
		log.Printf("error checking status: %v", err)
		return nil, err
	}
	resp.Result()
	if obj.Count == 0 {
		tenant.Name = name
		tenant.Slug = Slugify(name)
		tenant.Tags = []Tag{apcTag, customerTag, jobberTag}
		return c.addTenant(tenant)
	} else {
		return c.GetTenant(obj.Results[0].ID)
	}
}

func (c *Client) addTenant(tenant *Tenant) (*Tenant, error) {
	type TenantReq struct {
		Tenant
		Group int `json:"group"`
	}
	req := &TenantReq{*tenant, 1}
	r := c.buildRequest().SetResult(tenant).SetBody(req)
	resp, err := r.Post(c.buildURL(tenantPath))
	if err != nil {
		log.Fatalf("error adding tenant %s: %v\n", tenant.Name, err)
	}
	if err = checkStatus(resp); err != nil {
		log.Fatalf("error checking status: %v", err)
	}
	log.Printf("  added tenant %d %s", tenant.ID, tenant.Name)
	return tenant, nil
}

func (c *Client) GetTenant(id int) (*Tenant, error) {
	log.Printf("  getting tennant %d\n", id)
	tenant := &Tenant{ID: id}
	r := c.buildRequest().SetResult(tenant).SetBody(tenant)
	r.SetPathParam("id", fmt.Sprint(id))
	resp, err := r.Post(c.buildURL(tenantPath + "{id}"))
	if err != nil {
		log.Fatalf("error getting tenant %s: %v\n", tenant.Name, err)
	}
	if err = checkStatus(resp); err != nil {
		log.Fatalf("error checking status: %v", err)
	}
	return tenant, nil

}

// SearchDeviceAndVM searches both the devices and virtualmachines
// endpoints for the given args.  Calls SearchDevices() and SearchVMs()
// to get the results.
//
// Args should be specified as
// key=value (eg. has_primary_ip=true)
func (c *Client) SearchDeviceAndVM(args ...string) ([]DeviceOrVM, error) {
	var devices []DeviceOrVM
	devices, err := c.SearchDevices(args...)
	if err != nil {
		return nil, err
	}
	vms, err := c.SearchVMs(args...)
	if err != nil {
		return nil, err
	}
	devices = append(devices, vms...)
	for _, dev := range devices {
		setDeviceCustomFields(&dev)
	}
	return devices, nil
}

// SearchDevices searches  the devices
// endpoint for the given args.  Args should be specified as
// key=value (eg. has_primary_ip=true)
func (c *Client) SearchDevices(args ...string) ([]DeviceOrVM, error) {
	return c.performDevVMsearch("device", args...)
}

// performDevVMsearch executes the search for devices or VMs
func (c *Client) performDevVMsearch(objectType string, args ...string) ([]DeviceOrVM, error) {
	var devices []DeviceOrVM
	obj := DeviceVMSearchResults{}
	r := c.buildRequest().SetResult(&obj)
	path := GetPathForModel(objectType)
	if path == "" {
		c.log.Error("could not determine the path for model %s", objectType)
		return devices, fmt.Errorf("could not determine the path for model %s", objectType)
	}
	queryArgs := buildQueryPath(args...)
	initalURL := c.buildURL(path+"/?%s", queryArgs)
	url := &initalURL
	for url != nil {
		resp, err := r.Get(*url)
		if err != nil {
			c.log.Error(fmt.Sprintf("error searching %s", r.URL), "err", err)
			return devices, err
		}
		if resp.IsError() {
			c.log.Error(fmt.Sprintf("%d searching %s", resp.StatusCode(), r.URL), "err", err)
			return devices, fmt.Errorf("%s: %s", resp.Error(), resp.Body())
		}
		devices = append(devices, obj.Results...)
		url = obj.Next
	}
	for _, dev := range devices {
		setDeviceCustomFields(&dev)
	}
	return devices, nil
}

// buildQueryPath concatenates a strgin of arguments onto a queryPath separated by &
func buildQueryPath(args ...string) string {
	var queryArgs string
	concat := ""
	for _, arg := range args {
		queryArgs = fmt.Sprintf("%s%s%s", queryArgs, concat, arg)
		concat = "&"
	}
	return queryArgs
}

func (c *Client) Search(objectType string, resultObj any, args ...string) error {
	path := GetPathForModel(objectType)
	if len(args) > 0 {
		path = path + "/?%s"
	}
	queryArgs := buildQueryPath(args...)
	url := c.buildURL(path, queryArgs)
	req := c.buildRequest().SetResult(resultObj)
	resp, err := req.Get(url)
	if err != nil {
		c.log.Error("error communicating with netbox", "method", "GET", "url", url, "error", err)
		return err
	}
	if resp.IsError() {
		c.log.Error("netbox returned an error response", "method", "GET", "url", url, "status", resp.StatusCode())
		return fmt.Errorf("response error: %v", resp.Error())
	}
	return nil
}

func (c *Client) GetByID(objectType string, resultObj interface{}, id int) (interface{}, error) {
	path := GetPathForModel(objectType)
	url := fmt.Sprintf("%s/%d", c.buildURL(path), id)
	return c.GetByURL(url, resultObj)
}

// GetByURL is useful for iterating through search results using the Next URL value.
// The response object can be a pointer or object and will be returned.
func (c *Client) GetByURL(url string, obj interface{}) (interface{}, error) {
	r := c.buildRequest().SetResult(obj)
	resp, err := r.Get(url)
	if err != nil {
		c.log.Error(fmt.Sprintf("error calling %s", r.URL), "err", err)
		return obj, err
	}
	if resp.IsError() {
		c.log.Error(fmt.Sprintf("%d calling %s", resp.StatusCode(), r.URL), "err", err)
		return obj, fmt.Errorf("%s: %s", resp.Error(), resp.Body())
	}
	return obj, err
}

func setDeviceCustomFields(dev *DeviceOrVM) {
	if mon, ok := dev.CustomFieldsMap["monitoring_id"]; ok {
		if monid, isInt := mon.(int); isInt {
			dev.CustomFields.MonitoringID = &monid
		}
	}
}
