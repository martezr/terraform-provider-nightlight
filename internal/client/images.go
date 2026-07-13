package client

import "fmt"

func (c *Client) ListImages() ([]Image, error) {
	var out []Image
	return out, c.request("GET", "/api/v1/images", nil, &out)
}

func (c *Client) GetImage(id string) (*Image, error) {
	var out Image
	err := c.request("GET", fmt.Sprintf("/api/v1/images/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type CreateImageRequest struct {
	Name            string                   `json:"name"`
	Description     string                   `json:"description,omitempty"`
	OperatingSystem string                   `json:"operatingSystem,omitempty"`
	Format          string                   `json:"format,omitempty"`
	SizeGB          float64                  `json:"sizeGB,omitempty"`
	DatastoreId     string                   `json:"datastoreId,omitempty"`
	SourceType      string                   `json:"sourceType,omitempty"`
	SourcePath      string                   `json:"sourcePath,omitempty"`
	SourceURL       string                   `json:"sourceUrl,omitempty"`
	FileName        string                   `json:"fileName,omitempty"`
	Tags            []map[string]interface{} `json:"tags"`
}

func (c *Client) CreateImage(in CreateImageRequest) (*Image, error) {
	var out Image
	err := c.request("POST", "/api/v1/images", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteImage(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/images/%s", id), nil, nil)
}
