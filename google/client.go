package google

import (
	"net/http"
)

const (
	baseURL   = "https://www.googleapis.com/youtube/v3"
	searchURL = baseURL + "/search"

	apiKey = "AIzaSyDpnichUaq3woYXzHsJ99ALR4u2JpaR984"
)

// Client is a client for interacting with the Google APIs
type Client struct {
	client *http.Client
}

func (c *Client) assureClient() {
	if c.client != nil {
		return
	}
	c.client = http.DefaultClient
}
