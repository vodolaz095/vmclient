package vmclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	endpoint    string
	headers     map[string]string
	hclient     *http.Client
	extraLabels string
	u           *url.URL
}

func (c *Client) Close(context.Context) (err error) {
	c.hclient.CloseIdleConnections()
	return nil
}

func New(ctx context.Context, cfg Config) (vmc *Client, err error) {
	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("error parsing endpoint: %w", err)
	}

	vmc = &Client{
		endpoint:    u.String(),
		headers:     cfg.Headers,
		extraLabels: cfg.ExtraLabels,
		u:           u,
	}
	if cfg.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if cfg.HttpClient != nil {
		vmc.hclient = cfg.HttpClient
	} else {
		vmc.hclient = &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
	}
	err = vmc.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error pinging endpoint: %w", err)
	}
	return vmc, nil
}
