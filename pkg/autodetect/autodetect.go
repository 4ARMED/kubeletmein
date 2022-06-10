package autodetect

import (
	"net/http"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

// Providers stores a map of Provider structs
type Providers map[string]Provider

// Provider stores the details of each cloud provider
type Provider struct {
	Path               string
	RequestHeader      map[string]string
	ResponseHeader     map[string]string
	ExpectedStatusCode int
	Method             string
}

// Client wraps http.Client so we can mock
type Client struct {
	hc        *http.Client
	providers Providers
}

var (
	metadataServerURL = "http://169.254.169.254"

	// PublicCloudProviders contains the details we use to check which one
	// we are on.
	PublicCloudProviders = Providers{
		"gke": Provider{
			Path:               "/",
			Method:             "GET",
			RequestHeader:      map[string]string{},
			ResponseHeader:     map[string]string{"Server": "Metadata Server for VM"},
			ExpectedStatusCode: http.StatusOK,
		},
		"do": Provider{
			Path:               "/metadata/v1/id",
			Method:             "GET",
			RequestHeader:      map[string]string{},
			ResponseHeader:     map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			ExpectedStatusCode: http.StatusOK,
		},
		"eks": Provider{
			Path:               "/",
			Method:             "GET",
			RequestHeader:      map[string]string{},
			ResponseHeader:     map[string]string{"Server": "EC2ws"},
			ExpectedStatusCode: http.StatusOK,
		},
		"eks-imdsv2": Provider{
			Path:               "/latest/api/token",
			Method:             "PUT",
			RequestHeader:      map[string]string{"X-aws-ec2-metadata-token-ttl-seconds": "21600"},
			ResponseHeader:     map[string]string{"Server": "EC2ws"},
			ExpectedStatusCode: http.StatusOK,
		},
	}
)

// New provides a client to the metadata service
func New(hc *http.Client, providers Providers) (*Client, error) {
	if providers == nil {
		providers = PublicCloudProviders
	}

	return &Client{hc: hc, providers: providers}, nil
}

// GetProvider attempts to calculate the public cloud provider
// we are running on
func (c *Client) GetProvider() string {

	logger.Debug("beginning autodetection...")

	for name, provider := range c.providers {
		logger.Debug("trying [%s]", name)
		result := checkProvider(c.hc, provider)
		if result {
			return name
		}
	}
	return ""
}

func checkProvider(hc *http.Client, provider Provider) bool {

	rq, err := http.NewRequest(provider.Method, metadataServerURL+provider.Path, nil)
	if err != nil {
		return false
	}

	for k, v := range provider.RequestHeader {
		rq.Header.Add(k, v)
	}

	rs, err := hc.Do(rq)
	if err != nil {
		return false
	}

	if rs.StatusCode != provider.ExpectedStatusCode {
		return false
	}

	for k, v := range provider.ResponseHeader {
		header := rs.Header.Get(k)
		if header != v {
			logger.Debug("header %s does not match expected value %s, got %s", k, v, header)
			return false
		}
	}

	// If we got here then everything matches
	return true
}
