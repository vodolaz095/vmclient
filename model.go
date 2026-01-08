package vmclient

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func labelsToString(labels map[string]string) string {
	var name string
	var elems []string
	var i = 0
	for k := range labels {
		if k == LabelForName {
			name = labels[k]
			continue
		}
		elems = append(elems, fmt.Sprintf("%s=%q", k, labels[k]))
		i++
	}
	sort.Slice(elems, func(i, j int) bool {
		return elems[i] < elems[j]
	})
	return name + "{" + strings.Join(elems, ",") + "}"
}

type Result struct {
	Value     float64
	Timestamp time.Time
}

type Instant struct {
	Result
	Labels map[string]string
}

func (i *Instant) Name() string {
	name, found := i.Labels[LabelForName]
	if found {
		return name
	}
	return ""
}

func (i *Instant) String() string {
	return labelsToString(i.Labels)
}

type Range struct {
	Labels map[string]string
	Values []Result
}

func (r *Range) Name() string {
	name, found := r.Labels[LabelForName]
	if found {
		return name
	}
	return ""
}

func (r *Range) String() string {
	return labelsToString(r.Labels)
}
