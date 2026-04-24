package subdl

import (
	"context"
	"fmt"
	"sync"

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
	httpClient *httpclient.Client // Internal HTTP client
	mu         sync.RWMutex       // Protects access to token and currentBaseUrl
	authToken  *string
	baseURL    string
}

func NewClient(c Config) (*Client, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("subdl API Key is required to access Sub DL")
	}

	client := Client{
		config:     c,
		httpClient: httpclient.New(SUBDLAPIURL, c.APIKey, "subg v1"),
		baseURL:    SUBDLAPIURL,
	}

	return &client, nil
}

func (c *Client) SearchSubtitles(ctx context.Context, params SubDLSearchParams) (*SubDLSearchResults, error) {
	var response SubDLSearchResults

	err := c.httpClient.Get(ctx, "/subtitles", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
