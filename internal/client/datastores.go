package client

import "fmt"

func (c *Client) ListDatastores() ([]Datastore, error) {
	var out []Datastore
	return out, c.request("GET", "/api/v1/datastores", nil, &out)
}

func (c *Client) GetDatastore(id string) (*Datastore, error) {
	var out Datastore
	err := c.request("GET", fmt.Sprintf("/api/v1/datastores/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateDatastore(in Datastore) (*Datastore, error) {
	var out Datastore
	err := c.request("POST", "/api/v1/datastores", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDatastore(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/datastores/%s", id), nil, nil)
}
