package subdl

import (
	"context"
	"fmt"

	"github.com/kakeetopius/subg/internal/httpclient"
)

const (
	SUBDLAPIURL      = "https://api.subdl.com/api/v1"
	SUBDLDOWNLOADURL = "https://dl.subdl.com"
)

type Config struct {
	APIKey string
}

type Client struct {
	config     Config
	httpClient *httpclient.Client
	baseURL    string
}

func NewClient(c Config) (*Client, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("subdl API Key is required to access Sub DL")
	}

	client := Client{
		config:  c,
		baseURL: SUBDLAPIURL,
	}

	httpClient := httpclient.New(SUBDLAPIURL)
	httpClient.SetAPIKey(&c.APIKey)

	client.httpClient = httpClient
	return &client, nil
}

func (c *Client) SearchSubtitles(ctx context.Context, params SearchParams) (*SearchResults, error) {
	var response SearchResults

	err := c.httpClient.Get(ctx, "/subtitles", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
