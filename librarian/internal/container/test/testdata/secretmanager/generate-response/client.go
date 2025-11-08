// Package secretmanager provides a client for the secretmanager API.
package secretmanager

// Client is a client for the API.
type Client struct {
	// This is a test container placeholder.
}

// NewClient creates a new client.
func NewClient() *Client {
	return &Client{}
}

// V1Service provides access to the google/cloud/secretmanager/v1 API.
func (c *Client) V1Service() *V1Service {
	return &V1Service{}
}

// V1Service is a placeholder service.
type V1Service struct{}

// V1beta2Service provides access to the google/cloud/secretmanager/v1beta2 API.
func (c *Client) V1beta2Service() *V1beta2Service {
	return &V1beta2Service{}
}

// V1beta2Service is a placeholder service.
type V1beta2Service struct{}

