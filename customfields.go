package netbox

import (
	"errors"
	"fmt"
)

func (c *Client) UpdateCustomFieldOnModel(model string, modelID int64, field string, value any) error {
	cf := make(map[string]interface{})
	data := make(map[string]interface{})
	cf[field] = value
	data["custom_fields"] = cf

	return c.UpdateObjectWithMap(model, modelID, data)
}

// CustomFieldExists checks to see if a custom field exists in Netbox
// error is set if there's a problem communicating with Netbox
func (c *Client) CustomFieldExists(name string) (bool, error) {
	exists := false
	field := make(map[string]interface{})

	err := c.Search("customfield", &field, fmt.Sprintf("name=%s", name))
	if err != nil {
		return exists, err
	}
	count := field["count"].(float64)
	switch count {
	case 0:
		exists = false
	case 1:
		exists = true
	default:
		exists = true
		err = errors.New("too many results returned")
	}
	return exists, err
}

// AddCustomField adds the given name as a custom field.
//
//	name is the internal field name
//	label is the the display name
//	readonly indicates if the field should be editable
//	objects are the types of objects to attach the field to (at least 1 is requred)
func (c *Client) AddCustomField(name string, label string, readonly bool, objects ...string) error {
	data := make(map[string]interface{})
	data["name"] = name
	data["label"] = label
	data["type"] = "text"
	objs := []string{}

	if len(objects) == 0 {
		return errors.New("at least 1 object type must be specified")
	}
	for _, obj := range objects {
		objs = append(objs, getObjectType(obj))
	}
	data["content_types"] = objs
	if readonly {
		data["ui_editable"] = "no"
	}
	path := GetPathForModel("customfield") + "/"
	r := c.buildRequest()
	url := c.buildURL(path)
	r.SetBody(data)
	resp, err := r.Post(url)
	if err != nil {
		c.log.Error("could not add custom field", "field", name, "error", err)
		return err
	}
	if resp.IsError() {
		c.log.Error("netbox returned an error", "status", resp.StatusCode(), "body", resp.Body())
		return fmt.Errorf("%s: %s", resp.Error(), resp.Body())
	}
	return nil
}
