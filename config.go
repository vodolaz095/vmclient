package vmclient

import "net/http"

// Config defines connection parameters
type Config struct {
	Address     string
	Headers     map[string]string
	ExtraLabels string
	HttpClient  *http.Client
}
