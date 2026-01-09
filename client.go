package vmclient

import (
	"context"
	"crypto/tls"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	endpoint    string
	headers     map[string]string
	hclient     *http.Client
	extraLabels string
}

func (c *Client) Close(context.Context) (err error) {
	c.hclient.CloseIdleConnections()
	return nil
}

func New(ctx context.Context, cfg Config) (vmc *Client, err error) {
	vmc = &Client{
		endpoint:    cfg.Address,
		headers:     cfg.Headers,
		extraLabels: cfg.ExtraLabels,
	}
	if cfg.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if cfg.HttpClient != nil {
		vmc.hclient = cfg.HttpClient
	} else {
		vmc.hclient = otelhttp.DefaultClient
	}
	err = vmc.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return vmc, nil
}
