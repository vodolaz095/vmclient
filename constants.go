package vmclient

import (
	"time"
)

const LabelForName = "__name__"

const DefaultStep = 5 * time.Minute

const DefaultPushEndpoint = "/api/v1/import/prometheus"

const DefaultEndpoint = "http://127.0.0.1:8428"
