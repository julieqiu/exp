// Package accessapproval provides a client for the accessapproval API.
package accessapproval

// Client is a client for the API.
type Client struct {
	// This is a test container placeholder.
}

// NewClient creates a new client.
func NewClient() *Client {
	return &Client{}
}

// V1Service provides access to the google/cloud/accessapproval/v1 API.
func (c *Client) V1Service() *V1Service {
	return &V1Service{}
}

// V1Service is a placeholder service.
type V1Service struct{}

