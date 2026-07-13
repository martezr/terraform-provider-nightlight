package client

import "fmt"

func (c *Client) ListContentLibraries() ([]ContentLibrary, error) {
	var out []ContentLibrary
	return out, c.request("GET", "/api/v1/content-libraries", nil, &out)
}

func (c *Client) GetContentLibrary(id string) (*ContentLibrary, error) {
	var out ContentLibrary
	err := c.request("GET", fmt.Sprintf("/api/v1/content-libraries/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateContentLibrary(in ContentLibrary) (*ContentLibrary, error) {
	var out ContentLibrary
	err := c.request("POST", "/api/v1/content-libraries", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateContentLibrary(id string, in UpdateContentLibraryRequest) (*ContentLibrary, error) {
	var out ContentLibrary
	err := c.request("PUT", fmt.Sprintf("/api/v1/content-libraries/%s", id), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteContentLibrary(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/content-libraries/%s", id), nil, nil)
}
